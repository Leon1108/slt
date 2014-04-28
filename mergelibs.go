package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	WORK_DIR_NAME = "target" // 工作目录根目录名
)

var targetArchMap map[string]string = make(map[string]string)

// 合并静态库
func merge(dest string, libs []libraryInfo) bool {
	// 创建临时目录根目录
	workRootPath := createTempDir("", "")
	workDir := createTempDir(workRootPath, WORK_DIR_NAME)
	defer cleanTempDir(workRootPath)

	// 遍历所有静态库
	for _, lib := range libs {
		// 为每个静态库，创建一个文件夹
		libTmpPath := createTempDir(workRootPath, lib.path)
		defer cleanTempDir(libTmpPath)

		// 遍历所有CPU架构，抽取出每个CPU架构，保存为独立的静态库
		if len(lib.archs) == 1 {
			// 如果只有一个架构则直接将文件Copy到临时目录
			arch := lib.archs[0]
			archTmpPath := createTempDir(libTmpPath, arch)
			archWorkDir := buildTargetArchMap(workDir, arch)
			archTmpLib := archTmpPath + "/" + arch + ".a"
			syncExec("cp", lib.path, archTmpLib)
			extract(lib.absolutePath, archTmpPath)
			copyAll(archTmpPath, archWorkDir, "*.o")
		} else {
			for _, arch := range lib.archs {
				archTmpPath := createTempDir(libTmpPath, arch)
				archWorkDir := buildTargetArchMap(workDir, arch)
				archTmpLib := archTmpPath + "/" + arch + ".a"
				unarchive(lib.path, archTmpLib, arch)
				extract(archTmpLib, archTmpPath)
				copyAll(archTmpPath, archWorkDir, "*.o")
			}
		}
	}

	// libtool 生成静态库
	thins := make([]string, len(targetArchMap))
	for arch, path := range targetArchMap {
		thinOutput := path + "/" + arch + ".a"
		thins = append(thins, thinOutput)
		libtool(path, thinOutput)
	}

	// lipo 合并静态库
	lipoCreate(strings.Join(thins, " "), dest)
	return true
}

// 构造一个维护每个架构所对应的目标目录的字典
func buildTargetArchMap(workDir, arch string) string {
	archWorkDir := createTempDir(workDir, arch)
	if _, ok := targetArchMap[arch]; !ok {
		targetArchMap[arch] = archWorkDir
	}
	return archWorkDir
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

// 将所以符合<pattern>的文件从<src>目录拷贝到<dest>目录
func copyAll(src, dest, pattern string) {
	_, err := syncExec("/bin/sh", "-c", fmt.Sprintf("cp %v/%v %v", src, pattern, dest))
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

func cleanTempDir(path string) (err error) {
	err = os.RemoveAll(path)
	if nil != err {
		debug("Fail to remote template directory '%v'", path)
	} else {
		debug("Remove templage directory '%v'", path)
	}
	return
}
