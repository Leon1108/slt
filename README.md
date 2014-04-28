# 多架构静态库合并工具

## 限制：
+ 输入文件必须包含相同的架构,比如都只包含armv7, armv7s, i386。否则无法合并。
+ 当前版本仅仅是简单的将同名文件（.o文件）去重，如果两个静态库中包含的`文件版本不同`，则可能可能在运行时出现问题。

## 用法:
    mergelibs [-dhov] <input_files>
    输入文件数(input_files)需大于等于2个

    -d: 打印调试信息
    -h: 打印帮助信息
    -o <output>: 指定输出文件名称，默认会在执行命令的目录生成一个名为“merged.a”的文件
    -v: 打印版本信息

## 例如：
 1. $mergelibs -h
 2. $mergelibs -v
 3. $mergelibs xxx.a yyy.a
 4. $mergelibs -o all_in_one.a xxx.a yyy.a
 5. $mergelibs -d -o all_in_one.a xxx.a yyy.a