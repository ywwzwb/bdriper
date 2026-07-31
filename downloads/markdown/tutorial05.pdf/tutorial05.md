## VCB-Studio 教程 05: 封装、分离和转封装容器

在教程 4中，我们已经介绍了如何用 MeGUI自带的工具做最基本 MP4/MKV 的封装操作。本教程旨在继续介绍关于封装格式、分拆和转换，按照如下模块组织：

1、关于封装容器的一些知识

2、用 MKVtoolnix 封装 MKV

3、用 MeGUI 自带的 HD Stream Extractor 拆分 M2TS/MKV

4、用 MKVtoolnix 拆分 MKV

5、拆分MP4的办法

6、如何转封装格式

7、将多段视频/音频合并为一个

## 1、关于封装容器的一些知识

封装容器，就是我们日常见到的 mp4/mkv 等，在使用上有略微的区别。就我们日常接触的格式来说：

1. 视频方面，8bit/10bit AVC/HEVC两者都支持，因此没有太多区别。注意，视频本身可能自带封装，比如有些视频轨道（比如 x264/x265 编码可选后缀名）就是 mkv/mp4 后缀的。你把 mkv/mp4 的纯视频轨道扔给MKV 封装器没问题，但是你把 MKV的视频丢给MP4 封装器它就不认。这时候需要先把纯视频流从MKV容器中抽取出来，才能继续封装成 MP4。

2. 音频方面，MKV 基本全支持，MP4 不能支持 FLAC 音轨。

3. 字幕方面，MKV 支持 ass/ssa 字幕，以及原盘 PGS(后缀为.sup)字幕，MP4 以上都不支持。

MP4 的兼容性一般非常好，毕竟作为工业标准。几乎所有的工业性非编软件，以及播放器，都支持MP4封装。但是需要注意，能被支持的前提是，里面的视频也被支持。比如只能支持8bit AVC的播放器，你丢10bit，或者x265，显然是没用的。

不同时代，封装容器的规范一般有小的区别，比如10年前的MP4 不支持H.265，不支持ALAC，这是很容易理解的。如果你碰到你封装的东西，设备不支持，你可以选择找一版上古封装工具试试。

x264，以及Yukki版 x265，封装出来的MKV并不是最规范的，表现为播放的时候往往不能快进，甚至有花屏。解决方案是再封装一次。

在封装的时候，封装工具会对整个文件做一个扫描，记录下所有IDR帧等关键帧的位置，记录在文件头中，方便在播放的时候快速查找。x264/x265编码过程中，后续的段落未知，所以没办法做这个扫描，自然播放时候经常跳转不能了。

ffmpeg 往往能一键转封装，但是转出的成品不一定标准规范，又缺乏好用的 GUI（可能 MediaCoder 算一个），因此不推荐作为日常转封装使用，只适合在测试过成品的前提下，大规模批处理转任务。

## 2、用 MKVtoolnix 封装 MKV

MKVtoolnix 是一套全能的，处理 mkv 的工具。mkvtoolnix-gui 就是 mkvtoolnix 的主执行程序：

<!-- image-->

## 3、用 MeGUI 自带的 HD Stream Extractor 拆分 M2TS/MKV

MeGUI 自带的拆分器其实是个非常全能的，拆分mkv/m2ts 的工具，依赖的eac3to。教程4 中介绍的AutoEAC 其实是它的批量简化版，专职抽音轨。

<!-- image-->

如果是多音轨，一般全部抽取，然后保留不重复的。

除了抽取音轨和字幕外，它还可以用来抽取视频。最常见的就是从 AVC 的 MKV 中抽取 h264 文件（Extractas 需要选择 h264），用于封装 MP4。注意这玩意尚不支持从 HEVC 的 MKV 中抽 h265 文件。

## 4、用 MKVtoolnix 拆分 MKV

mkvmerge 文件夹下，还有个 MKVExtractorGUI2.exe。这东西是 mkv 拆分工具的图形界面，使用方式很

<!-- image-->

这东西支持从 HEVC 的 MKV 中抽取 H265 视频流。抽取出来的东西可能需要手动加一下后缀.265，才能被 MeGUI 的 MP4 封装工具识别。

注意章节抽取出来后缀名称是 ogm，直接改 txt 就行了。

## 5、拆分 MP4 的办法

MP4 的拆分需要借助一个叫做 Yamb MP4 tools 的工具，不过这个工具实在是太反人类了（有兴趣的可以自己去下了玩玩），所以实际操作推荐大家先把 mp4 丢 mmg.exe 封装成 mkv，再当做 MKV 去拆。

## 6、重封装的方法

任何格式转 MKV 封装，直接将源文件拖入 mkvtoolnix-gui.exe 就好了，注意设置轨道细节；

任何格式转 MP4 封装，先抽取所有轨道的 raw 格式（264，AAC，……），再用 MeGUI 封装。

其他非常规手段包括使用 ffmpeg（命令行工具）:

ffmpeg -i a.mkv -vcodec copy -acodec copy a.mp4

就是将 a.mkv 转为 a.mp4

或者使用 MediaCoder，复制视频流和音频流，选不同的容器。

注意，不同工具封装的结果，兼容性和规范性不一定相同，哪怕后缀是一样的。常规方法可以保证比较安全的兼容性和规范性。

## 7、将多段视频和音频合并为一个

一个简单的方式是使用mkvtoolnix 合并。注意合并的前提不但是参数一样，还要求编码参数一样。前后两段必须是使用同样的 x264 参数编码而成。

mkvtoolnix 连接文件，采用的是追加合并（append）按钮，在老版本的 mmg.exe 上如下操作：

<!-- image-->

完成之后，一定要播放检查，播放效果是否正常，特别是连接的位置。如果出现下图所示的报错信息，说明两个轨道前后格式不匹配：

<!-- image-->

新版本的 mkvtoolnix-gui 操作逻辑是一样的：

<!-- image-->

如果有大批零散文件，常见于那种电影原盘，正片分散在许多 m2ts中，需要抽取音轨或视频，这时候可以使用 MeGUI 的 HD Stream Extractor 来处理。HD Stream Extractor 可以多选，按住 ctrl 先后选择多个

文件后，MeGUI 会自动把它们当做合并的文件对待，抽取音轨啥的可以一步到位：

MeGUI - HD-DVD/Blu-ray Streams Extractor

-Feature(s)

<table><tr><td># Name</td><td>Description</td><td>File(s)</td><td></td><td>Duration</td></tr><tr><td>0 00005.m2ts</td><td>00005.m2ts, 00:14:16</td><td>00005.m2ts</td><td></td><td>00:14:16</td></tr><tr><td></td><td></td><td></td><td></td><td></td></tr></table>

Stream(s) 如果你选的前后格式一致，下面的选项会一会儿后弹出，否则不会。

<table><tr><td rowspan=1 colspan=1>Extract?</td><td rowspan=1 colspan=1>#</td><td rowspan=1 colspan=1>Type</td><td rowspan=1 colspan=1>Description</td><td rowspan=1 colspan=2>Extract As</td><td rowspan=1 colspan=1>+ Options</td><td rowspan=1 colspan=1> Numb</td></tr><tr><td rowspan=1 colspan=1>√</td><td rowspan=1 colspan=1>1</td><td rowspan=1 colspan=1>Video</td><td rowspan=1 colspan=1>h264/AVC, 1080i60 /1.001 (16:9)</td><td rowspan=1 colspan=2>MKV</td><td rowspan=1 colspan=1></td><td rowspan=1 colspan=1>1</td></tr><tr><td rowspan=1 colspan=1>√</td><td rowspan=1 colspan=1>2</td><td rowspan=1 colspan=1>Audio</td><td rowspan=1 colspan=1> RAW/PCM, 2.0 channels, 24 bits, 48kHz</td><td rowspan=1 colspan=2>FLAC</td><td rowspan=1 colspan=1></td><td rowspan=1 colspan=1>2</td></tr><tr><td rowspan=1 colspan=1>√</td><td rowspan=1 colspan=1>3</td><td rowspan=1 colspan=1>Audio</td><td rowspan=1 colspan=1> RAW/PCM, 2.0 channels, 24 bits, 48kHz</td><td rowspan=1 colspan=1>AC3</td><td rowspan=1 colspan=1></td><td rowspan=1 colspan=1></td><td rowspan=1 colspan=1>3</td></tr></table>

FeatureRetrievalCompleted到上面的流，说明选的文件格式不匹配，不能连接