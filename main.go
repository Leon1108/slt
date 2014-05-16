package main

import (
	"os"
	"fmt"
	"flag"
	"sort"
	"errors"
	"strings"
    "text/template"
)

type AppInfo struct {
    Command, Version, XcodeDefault, OutputDefault string
}

type LibraryInfo struct {
	path         string   // 静态库文件所在路径
	absolutePath string   // 静态库文件所在绝对路径
	archs        []string // 静态库文件所包含的CPU架构
}

const (
	CMD_NAME = "slt"
	VERSION	= "0.3.1"
	USAGE_TPL = `
{{.Command}} {{.Version}} -- Static Library Tools
============================================================================
Usage:
    {{.Command}} [-mpdhov] <input_files>

    -m: 工作模式
        merge       合并多架构静态库。[默认]
        exclude     排除指定文件。
    -p: Pattern 用于指定需要排除哪些文件。当工作模式为exclude时，该参数有效。
    -d: 打印调试信息
    -h: 打印帮助信息
    -o <output>: 指定输出文件名称，默认会在执行命令的目录生成一个名为'{{.OutputDefault}}'的文件。
    -v: 打印版本信息
    -e: 指定Xcode安装位置，默认为'{{.XcodeDefault}}'

Example：
    [1] {{.Command}} -h
    [2] {{.Command}} -v
    [3] {{.Command}} xxx.a yyy.a
    [4] {{.Command}} -o all_in_one.a xxx.a yyy.a
    [5] {{.Command}} -d -o all_in_one.a xxx.a yyy.a
    [6] {{.Command}} -m exclude -p 'Pods.*-dummy.o' -o excluded.a xxx.a
============================================================================
`
	FLAG_OUTPUT_DEFAULT = "slt-output.a"
	FLAG_DEBUG_DEFAULT  = false
	FLAG_VERSION_DEFAULT = false
	FLAG_HELP_DEFAULT = false
    FLAG_ENV_DEFAULT = "/Applications/Xcode.app/"
)

const (
    MODE_MERGE = "merge"
    MODE_EXCLUDE = "exclude"
)

var flagOutput string
var flagDebug bool = FLAG_DEBUG_DEFAULT
var flagVersion bool = FLAG_VERSION_DEFAULT
var flagHelp bool = FLAG_HELP_DEFAULT
var flagWorkMode string
var flagPattern string
var flagEnv string  // Xcode 安装位置

var workMode string = MODE_MERGE    // 当前的工作模式
var xcodeCmdPath string     // Xcode命令行工具根目录
var libs []LibraryInfo

func init() {
	flag.StringVar(&flagOutput, "o", FLAG_OUTPUT_DEFAULT, "Output static file")
    flag.StringVar(&flagWorkMode, "m", MODE_MERGE, "Work mode")
    flag.StringVar(&flagPattern, "p", "", "Pattern")
	flag.BoolVar(&flagDebug, "d", FLAG_DEBUG_DEFAULT, "Output debugging information")
	flag.BoolVar(&flagVersion, "v", FLAG_VERSION_DEFAULT, "Print version info")
	flag.BoolVar(&flagHelp, "h", FLAG_HELP_DEFAULT, "Print useage")
    flag.StringVar(&flagEnv, "e", FLAG_ENV_DEFAULT, "The root path where you install your Xcode.")
}

//
// $slt -o output.a intput_1.a intput_2.a input_3.a
//
func main() {
    //  解析输入参数
	flag.Parse()

    // 如果包含 -h 参数则直接打印帮助信息，不论是否包含其他参数
	if flagHelp {
		printUsage()
		return
	}

    // 如果包含 -v 参数则直接打印版本信息，不论是否还包含其他参数
	if flagVersion {
		printVersionInfo()
		return
	}

    // 获取工作模式
    switch flagWorkMode {
        case MODE_EXCLUDE:
            workMode = MODE_EXCLUDE
        case MODE_MERGE:
            workMode = MODE_MERGE
        default:
            // ERROR 未知的工作模式
            panic(fmt.Sprintf("Unknown work mode!! '%v'", flagWorkMode))
    }
    log("SLT (Static Library Tools) work in [%v] mode.", workMode)

    // 检查工作环境
    xcodeCmdPath = flagEnv + CMDBASE
    log("Commonds Path: %v", xcodeCmdPath)
    if !checkEnvironment() {
        panic("当前系统环境不支持此工具使用，请确保正确安装Xcode！如果Xcode没有安装在'/Applications' 目录下，请使用'-e'选项指定Xcode安装位置，如：\n\t$slt -e '/Users/UserName/Desktop/Xcode.app/' libXXX.a libYYY.a\n")
    }

	// 读取其余输入参数，也就是，非flag部分的参数，一般是输入文件
	inputFiles := flag.Args()

	// 检查input_files
	if !checkInputFiles(inputFiles) {
		return
	}

    // 根据所处的工作模式，开始相应的处理工作
    switch workMode {
    case MODE_MERGE:
	    // 开始合并工作
        if merge(libs, flagPattern, flagOutput) {
            log("Success! Save to %v", flagOutput)
        } else {
            log("Failed!")
        }
    case MODE_EXCLUDE:
        // 参数检查
        if len(flagPattern) > 0 {
            if merge(libs, flagPattern, flagOutput) {
                log("Success! Save to %v", flagOutput)
            } else {
                log("Failed!")
            }
        } else {
            panic("错误！请指定要被剔除的文件需要满足的模式，通过'-p'选项。")
        }
    }
}

// 检查当前运行环境，主要是检查Xcode是否正确安装
func checkEnvironment() bool{
    // 检查Xcode根目录是否存在
    if !IsDirExist(flagEnv) {
        errLog("指定的目录不存在！'%v'", flagEnv)
        return false;
    }

    // 检查所有命令行工具是否存在
    cmds := []string{CMD_AR, CMD_LIPO, CMD_LIBTOOL, CMD_OTOOL}
    for _, v := range cmds {
        cmdPath := xcodeCmdPath + v
        if !IsFileExist(cmdPath) {
            errLog("命令行工具'%v'没有找到！", cmdPath)
            return false
        } else {
            debug("Command: [%v] Founded!", cmdPath)
        }
    }
    return true;
}

// 验证输入文件可用性
func checkInputFiles(inputs []string) bool {

	// 检查输入文件中是否包含相同的文件
	tmp := make(map[string]string)
	for _, v := range inputs {
		_, ok := tmp[v]
		if ok {
			errLog("错误: 输入中包含相同的文件！%v", v)
			return false
		} else {
			tmp[v] = ""
		}
	}

	// 检查输入文件个数
    switch len(inputs){
    case 0:
		errLog("错误: 没有输入文件")
		printUsage()
		return false
    case 1:
        if workMode == MODE_MERGE {
		    errLog("错误: 没有足够的输入文件。在merge模式下需要至少2个输入文件。")
            printUsage()
            return false
        }
    default:
        if workMode == MODE_EXCLUDE{
		    errLog("错误: 输入文件过多。在exclude模式下仅支持1个输入文件。")
            printUsage()
            return false
        }
    }

	// 检查输入文件有效性
	for _, file := range inputs {
		_, err := os.Open(file)
		if nil != err {
			// 打开文件失败
			errLog("错误: 无法打开输入文件 '%v'", file)
			return false
		} else {
			// 检查输入的文件是否拥有相同的架构
			archs, errArch := checkArchitecture(file)
			if nil != errArch {
				errLog("错误：输入文件格式错误！'%v' Error: %v", file, errArch)
				return false
			} else {
				// 构建library infos
				absPath := file
				pwd, _ := os.Getwd()
				// 简单处理一下路径格式
				if !strings.HasPrefix(file, "/") {
					if strings.HasPrefix(file, "./") {
						file = file[2:]
					}
					absPath = pwd + "/" + file
				}

				libInfo := LibraryInfo{
					path:         strings.Replace(file, " ", "\\ ", -1),
					absolutePath: strings.Replace(absPath, " ", "\\ ", -1),
					archs:        archs,
				}
				libs = append(libs, libInfo)
			}
		}
	}

	// 检出输入的多个静态库文件，是否拥有相同的架构
	joind := ""
	for _, lib := range libs {
		tmp := strings.Join(lib.archs, "")
		if len(joind) == 0 {
			joind = tmp
			continue
		}
		if joind != tmp {
			// 架构不一致
			errLog("错误：输入文件所包含的CPU架构不一致!")
			for k, v := range libs {
				errLog("\t[%v] %v: %v", k+1, v.path, v.archs)
			}
			return false
		}
	}

	return true
}

// 检查输入文件是否是有效的静态库文件
func checkArchitecture(file string) (archInfos []string, err error) {
	debug("Analyse %v ......", file)
	// 检查输入的文件是否为静态库文件
	if !isStaticLabrary(file) {
		err = errors.New("输入文件不是合法的静态库文件!")
		return
	}

	// 调用Lipo
	out, err := syncExec("/bin/sh", "-c", fmt.Sprintf("%v -info %v", getCommandPath(CMD_LIPO), file))
	if nil != err {
		return
	}

	sections := strings.Split(out, ":")
	archLine := sections[len(sections)-1]
	archLine = strings.TrimSpace(archLine)
	archInfos = strings.Split(archLine, " ")
	sort.Strings(archInfos)
	return
}

// 打印使用方法
func printUsage() {
    appInfo := AppInfo{CMD_NAME, VERSION, FLAG_ENV_DEFAULT, FLAG_OUTPUT_DEFAULT}
    tmpl, err := template.New("usage").Parse(USAGE_TPL)
    if nil != err {panic(err)}
    err = tmpl.Execute(os.Stdout, appInfo)
    if nil != err {panic(err)}
}

// 打印版本信息
func printVersionInfo(){
	fmt.Println(VERSION)
}

// 日志
func log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func errLog(format string, args ...interface{}) {
	fmt.Printf("ERROR -> " + format + "\n", args...)
}

func debug(format string, args ...interface{}) {
	if flagDebug {
		log("DEBUG -> "+format, args...)
	}
}

