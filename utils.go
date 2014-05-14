package main

import(
	"os"
    "fmt"
	"strconv"
	"time"
)

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

// 将每个CPU架构抽取出来
func unarchive(src, target, arch string) {
	_, err := syncExec("lipo", "-thin", arch, "-o", target, src)
	if nil != err {
		panic(err) // TODO
	}
}

// 抽取初静态库中的.o文件
func extract(srcLib, targetDir string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("cd %v && ar -x %v", targetDir, srcLib))
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
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("libtool -static -o %v %v/*.o", output, src))
	if nil != err {
		panic(err)
	}
}

func lipoCreate(src, output string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("lipo -create %v -output %v", src, output))
	if nil != err {
		panic(err)
	}
}

