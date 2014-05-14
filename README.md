# 静态库操作工具

## 限制：
+ 输入文件必须包含相同的架构,比如都只包含armv7, armv7s, i386。否则无法合并。
+ 当前版本仅仅是简单的将同名文件（.o文件）去重，如果两个静态库中包含的`文件版本不同`，则可能可能在运行时出现问题。

## 安装
1. 通过`go get`
    * 首先确保正确安装了golang [参考](http://golang.org/doc/install#install)
    * 正确设置环境变量 `$GOROOT` `$GOPATH`，并已将 `$GOPATH/bin` 放入 `$PATH` 中。
    * 通过 `go get -u github.com/Leon1108/slt` 即可安装、更新了。
2. 直接获取对应平台的可执行文件
	* TODO

## 用法:
    slt [-mpdhov] <input_files>

    -m: 工作模式
        merge       合并多架构静态库。[默认]
        exclude     排除指定文件。
    -p: Pattern 用于指定需要排除哪些文件。当工作模式为exclude时，该参数有效。
    -d: 打印调试信息
    -h: 打印帮助信息
    -o <output>: 指定输出文件名称，默认会在执行命令的目录生成一个名为“merged.a”的文件
    -v: 打印版本信息

## 例如：
 1. $slt -h
 2. $slt -v
 3. $slt xxx.a yyy.a
 4. $slt -o all_in_one.a xxx.a yyy.a
 5. $slt -d -o all_in_one.a xxx.a yyy.a
 6. $slt -m exclude -p 'Pods.*-dummy.o' -o excluded.a xxx.a
