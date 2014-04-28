package main

import (
	"os"
	"fmt"
	"flag"
	"sort"
	"bytes"
	"errors"
	"strings"
	"os/exec"
    "text/template"
)

type AppInfo struct {
    Command, Version string
}

type LibraryInfo struct {
	path         string   // 静态库文件所在路径
	absolutePath string   // 静态库文件所在绝对路径
	archs        []string // 静态库文件所包含的CPU架构
}

const (
	CMD_NAME = "slt"
	VERSION	= "0.1.0"
	USAGE_TPL = `
{{.Command}} {{.Version}} -- Multi-architecture static library tools
============================================================================
Usage:
    {{.Command}} [-dhov] <input_files>
    输入文件数(input_files)需大于等于2个

    -d: 打印调试信息
    -h: 打印帮助信息
    -o <output>: 指定输出文件名称，默认会在执行命令的目录生成一个名为“merged.a”的文件
    -v: 打印版本信息

Example：
    [1] {{.Command}} -h
    [2] {{.Command}} -v
    [3] {{.Command}} xxx.a yyy.a
    [4] {{.Command}} -o all_in_one.a xxx.a yyy.a
    [5] {{.Command}} -d -o all_in_one.a xxx.a yyy.a
============================================================================
`
	FLAG_OUTPUT_DEFAULT = "merged.a"
	FLAG_DEBUG_DEFAULT  = false
	FLAG_VERSION_DEFAULT = false
	FLAG_HELP_DEFAULT = false
)

var flagMergedOutput string
var flagDebug bool = FLAG_DEBUG_DEFAULT
var flagVersion bool = FLAG_VERSION_DEFAULT
var flagHelp bool = FLAG_HELP_DEFAULT
var libs []LibraryInfo

func init() {
	flag.StringVar(&flagMergedOutput, "o", FLAG_OUTPUT_DEFAULT, "Output static file")
	flag.BoolVar(&flagDebug, "d", FLAG_DEBUG_DEFAULT, "Output debugging information")
	flag.BoolVar(&flagVersion, "v", FLAG_VERSION_DEFAULT, "Print version info")
	flag.BoolVar(&flagHelp, "h", FLAG_HELP_DEFAULT, "Print useage")
}

//
// $slmt -o output.a intput_1.a intput_2.a input_3.a
//
func main() {
	flag.Parse()

	if flagHelp {
		printUsage()
		return
	}

	if flagVersion {
		printVersionInfo()
		return
	}

	// 读取其余输入参数
	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		printUsage()
		return
	}

	// 检查input_files
	if !checkInputFiles(inputFiles) {
		return
	}

	// 开始合并工作
	if merge(flagMergedOutput, libs) {
		log("Success! Save to %v", flagMergedOutput)
	} else {
		log("Failed!")
	}
}

// 验证输入文件可用性
func checkInputFiles(inputs []string) bool {

	// 检查输入文件中是否包含相同的文件
	tmp := make(map[string]string)
	for _, v := range inputs {
		_, ok := tmp[v]
		if ok {
			log("错误: 输入中包含相同的文件！%v", v)
			return false
		} else {
			tmp[v] = ""
		}
	}

	// 检查输入文件个数
	if len(inputs) <= 1 { // input file 必须大于等于2个
		log("错误: 至少需要包含2个输入文件")
		printUsage()
		return false
	}

	// 检查输入文件有效性
	for _, file := range inputs {
		_, err := os.Open(file)
		if nil != err {
			// 打开文件失败
			log("错误: 无法打开输入文件 '%v'", file)
			return false
		} else {
			// 检查输入的文件是否拥有相同的架构
			archs, errArch := checkArchitecture(file)
			if nil != errArch {
				log("错误：输入文件格式错误！'%v' Error: %v", file, errArch)
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
			log("错误：输入文件所包含的CPU架构不一致!")
			for k, v := range libs {
				log("      [%v] %v: %v", k+1, v.path, v.archs)
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
	out, err := syncExec("lipo", "-info", file)
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

func isStaticLabrary(file string) bool {
	out, err := syncExec("otool", "-f", file)
	if nil != err {
		return false
	}
	if len(out) == 0 {
		return false
	}
	return true
}

// 对Cmd.Run()的简单封装
func syncExec(command string, args ...string) (stdOutput string, err error) {
	var stdOut bytes.Buffer
	cmd := exec.Command(command, args...)
	debug("Exec: %v", cmd.Args)
	cmd.Stdout = &stdOut
	err = cmd.Run()
	if nil != err {
		return
	}
	stdOutput = string(stdOut.Bytes())
	return
}

// 打印使用方法
func printUsage() {
    appInfo := AppInfo{CMD_NAME, VERSION}
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

func debug(format string, args ...interface{}) {
	if flagDebug {
		log("DEBUG -> "+format, args...)
	}
}
