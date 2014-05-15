package main

import (
	"fmt"
	"os"
	"time"
	"bytes"
	"strconv"
	"os/exec"
)

const (
	CMDBASE = "/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin/"
	CMD_AR      = "ar"
	CMD_LIPO    = "lipo"
	CMD_OTOOL   = "otool"
	CMD_LIBTOOL = "libtool"
)

// 文件是否存在
func IsFileExist(path string) bool {
    return IsExist(path, false)
}

// 目录是否存在
func IsDirExist(path string) bool {
    return IsExist(path, true)
}

// 检查文件或目录是否存在
func IsExist(path string, isDir bool) bool {
    fi, err := os.Stat(path)
    var result bool
    if err != nil {
        result = os.IsExist(err)
    } else {
        if isDir {
            result = fi.IsDir()
        } else {
            result = !fi.IsDir()
        }
    }
    return result
}

// 创建临时文件夹
func createTempDir(parent, subdir string) (path string) {
	if len(parent) == 0 {
		path = os.TempDir() + strconv.Itoa(int(time.Now().UnixNano()))
	} else {
		path = parent + "/" + subdir
	}

	err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if nil != err {
		panic(err) // TODO
	} else {
		debug("Create template directory '%v'", path)
	}

	return
}

// 清理临时文件夹
func cleanTempDir(path string) (err error) {
	err = os.RemoveAll(path)
	if nil != err {
		debug("Fail to remote template directory '%v'", path)
	} else {
		debug("Remove templage directory '%v'", path)
	}
	return
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

// 获得命令的绝对路径
func getCommandPath(cmd string) string {
    return xcodeCmdPath + cmd
}

// 文件是否为静态库文件
func isStaticLabrary(file string) bool {
	out, err := syncExec("/bin/sh", "-c", fmt.Sprintf("%v -f %v", getCommandPath(CMD_OTOOL), file))
	if nil != err {
		return false
	}
	if len(out) == 0 {
		return false
	}
	return true
}

// 将每个CPU架构抽取出来
func unarchive(src, target, arch string) {
	_, err := syncExec("/bin/sh", fmt.Sprintf("%v -thin %v -o %v %v", getCommandPath(CMD_LIPO), arch, target, src))
	if nil != err {
		panic(err) // TODO
	}
}

// 抽取初静态库中的.o文件
func extract(srcLib, targetDir string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("cd %v && %v -x %v", targetDir, getCommandPath(CMD_AR), srcLib))
	if nil != err {
		panic(err) // TODO
	}
}

// 将所以符合<pattern>并且不符合<exclude>的文件从<src>目录拷贝到<dest>目录
func copyAll(src, dest, pattern, exclude string) {
	// 排除掉pods生成的 *dummy.o 文件
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("cp -rf `ls %v/%v | grep -E -v \"Pods-.*dummy.o\"` %v", src, pattern, dest))
	if nil != err {
		panic(err) // TODO
	}
}

func libtool(src, output string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("%v -static -o %v %v/*.o", getCommandPath(CMD_LIBTOOL), output, src))
	if nil != err {
		panic(err)
	}
}

func lipoCreate(src, output string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("%v -create %v -output %v", getCommandPath(CMD_LIPO), src, output))
	if nil != err {
		panic(err)
	}
}
