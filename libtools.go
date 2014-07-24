package main

import (
	"strings"
)

const (
	WORK_DIR_NAME = "target" // 工作目录根目录名
)

var targetArchMap map[string]string = make(map[string]string)

// 合并静态库
// 也可以通过该方法来过滤静态库中的文件，可以通过pattern参数来传入不希望出现在静态库中得文件（可以使用正则表达式）
func merge(libs []LibraryInfo, pattern, dest string) bool {
	debug("Start to merge ...... ")
    log("Libs:")
    for _, v := range libs {
        log("|-- %v", v)
    }
    log("Pattern: %v", pattern)
    log("Destination: %v", dest)

	// 创建临时目录根目录
	workRootPath := createTempDir("", "")
	workDir := createTempDir(workRootPath, WORK_DIR_NAME)
	defer cleanTempDir(workRootPath)

	// 遍历所有静态库
	for _, lib := range libs {
		// 为每个静态库，创建一个临时文件夹，存放临时文件
		libTmpPath := createTempDir(workRootPath, lib.path)
		defer cleanTempDir(libTmpPath)

        // 将此Fat静态库的所有架构分别解压并拷贝所有.o文件到工作目录
        extractAllArchs(lib, libTmpPath, workDir, pattern)
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

func extractAllArchs(lib LibraryInfo, libTmpPath, workDir, exclude string) {
	// 遍历所有CPU架构并解压到相应临时目录
	for _, arch := range lib.archs {
		archTmpPath := createTempDir(libTmpPath, arch)   // src
		archWorkDir := buildTargetArchMap(workDir, arch) // dest
		archTmpLib := archTmpPath + "/" + arch + ".a"

		if len(lib.archs) > 1 {
			unarchive(lib.path, archTmpLib, arch)
		} else {
			// 如果只包含一个架构则直接拷贝到临时目录中
			syncExec("cp", lib.path, archTmpLib)
		}

		extract(archTmpLib, archTmpPath)
		copyAll(archTmpPath, archWorkDir, "*.o", exclude)
	}
}

// 构造一个维护每个架构所对应的目标目录的字典
func buildTargetArchMap(workDir, arch string) string {
	archWorkDir := createTempDir(workDir, arch)
	if _, ok := targetArchMap[arch]; !ok {
		targetArchMap[arch] = archWorkDir
	}
	return archWorkDir
}

// vim: set expandtab ts=4 sts=4 sw=4 :
