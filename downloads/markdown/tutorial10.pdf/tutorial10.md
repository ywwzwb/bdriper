# VCB-Studio 教程 10 x265 v2.9 参数设置

本教程旨在讲述 x265 参数设计的技巧。编码器版本适用 x265 v2.9+8, 10bit 版本。

## 1. 参数分类

一般来说，编码器都有 document 来描述有哪些参数供你设置，这些参数大概是做什么的。测试的第一步就是先阅读 doc，根据你的经验，把参数分为这三类:

编码规范/specification, 这类参数一般是规定编码一些格式规范、编码器工作的。比如 x264/x265 中--profile--level --matrix, --display-window --sar 等等。这些参数一般无需测试，该怎样就怎样。一般这些参数的调整也不会显著影响编码速度和编码画质。

取舍性/trade off, 这类参数一般是时间换画质的，比如 x264/x265 中 --ref --bframes --me --subme--merange --rect --amp 等。这些参数对画质的影响往往是通用性的，不随片源类型、码率高低变化太多。

码率控制/rate control, 这类参数决定码率的分配，分配的多少，怎么个分配法。比如 x264/x265 中 --crf --qcomp--aq --psy 等等。这些参数对画质的影响往往体现在目视效果上，跑分并不能很好的体现，且不同类型片源、码率表现很不一样。

本教程将从这三个分类的角度，讲述 x265 一些参数的设置技巧。

## 2. 编码规范/specification

spec 相关的参数一般没多少需要手动指定的，一般只有这几个：

--depth 10/-D 10，表示输出精度，10 就是 10bit。如果已经使用了对应精度的 x265，这一项无需指定。

--no-open-gop，关闭 OpenGOP，屏蔽一些设备上不能正确解码 opengop 的问题。

--keyint 360 --min-keyint 1，GOP 区间长度。注意 HEVC 解码压力相比 AVC 较大，GOP 区间不宜设置太长。

--colormatrix bt709 --range limited, YUV 转 RGB 相关。

--deblock -1:-1，类似 x264.

## 3. 取舍性/trade off

--preset slower，preset 是官方给你准备的“一键设置”，因此当你不是很了解 x265 参数的时候，建议使用。

--ctu 32，表示最大允许 32x32 的 transform unit。虽然 x265 允许 64x64，但是过大的 TU 会增加平面的涂抹，增加运算量，降低多线程优化可能，总体来说在<=1080p 的编码下弊大于利。因此限制一下比较好。

--qg-size 8, 表示 qp 值调整的最低单位是 8x8 的 coding unit。这个数字越低，x265 调整一帧内 qp 值的灵活度越高。编码表现来看，固定其他参数（crf 模式下），这个值越低，画质越好，码率也越高。微调crf至同档码率，我们依旧认为--qg-size 8 的效果是最好的。

--me 3，使用 star search。star search 综合来说好于 umh，但是不要试图用 full。因为 x265 目前还没有对 full做优化，太慢了。

--subme 4，最低建议 3（preset = slow 时候自动设置为 3）。subme=3 开始在 ME 过程中考虑 chroma residual--me-range 38，官方建议 57，事实证明有点大。1080p 给 38 左右就绰绰有余了。

## --b-intra，允许 B 帧中出现 Intra Block。动画建议

--no-rect/--no-amp。rect 和 amp 是 HEVC 规范中对 block 的创新。一般来说 block 都是正方形，比如 32x32,8x8之类的。rect 启用 1x2/2x1 类型的，比如允许 32x16 的 block，8x16 的 block；amp 允许 1x4/3x4/4x1/4x3 类型的 block，比如 8x32,12x16 之类的。amp 的启用必须启用 rect，反之则不需要。--no-rect 表示不启用 rect（也顺道相当于--no-amp）。通常来说，<=1080p 下，rect 基本上没什么作用，amp 是几乎完全没作用，但是这俩都是速度黑洞。因此从效率角度建议关闭，或者至少关闭 amp

如果--rect 开启，建议开启--limit-tu 4 来限制 x265 对分块的（可能）无效尝试。

--ref 4。ref的意义和x264 中相似；不过实测ref 增加在x265 中作用不明显。建议不超过6

--weightb。允许 b 帧的加权预测，在一些渐变场景比较有用。

--bframes 8。b 帧并不是越高越好，建议给 6\~10 左右。

--rc-lookahead 60。编码时候往前看多少帧来规划 Coding Unit Tree（CUTree，相当于 MBTree）一般设置为60\~80 比较合理；帧率越高的片源适合给的越高。

--rd 3。Rate distortion optimization 的模式，越高，计算度越复杂。3 是一个比较平衡的选择。目前 5 是实际最高选择。根据官方 doc，--rd 3 和--rd 4 相同，--rd 5 和--rd 6 相同。

## 4. 码率控制/rate control

码率控制这块是 x265调节的重中之重。x265的威力只有配合高度定制化的参数才能真正显露出来。

--no-sao。 SAO 官方名称叫“Sample Adatpive Offset”，然而我一般称为“Smoothing All Objects”。sao 的启用虽然可以减少 DCT ringing 等欠码瑕疵，但是代价是极其暴力的涂抹效果。除非是极低码率编码，否则一般不推荐开启 SAO

--crf 18.0。 --crf依旧是调节体积/画质最有效，最直接的参数。默认的28.0大概是为了强调 x265在低码率的优势。日常编码怎么着 23以上吧。对于动漫的高画质编码，建议至少 18.0 起。

--aq-mode 2。x265 目前有三种 aq 模式。aq-mode 1 是最安全稳定的 aq，适合高码率/高画质编码；aq-mode2 相对来说效率最高，适合中低码率的编码；aq-mode 3对暗场进行加强，适合8bit编码防止暗场压烂。一般10bit编码根据 crf 高低决定 aq 选取，个人建议在 crf <= 16 时候使用 aq-mode 1，否则使用 aq-mode 2。注意同 crf下，不同 aq-mode 出来的体积是不一样的，3>1>2。

--aq-strength 0.9。aq-strength 决定了 aq 的强度，一般来说，动漫的 aq-strength 不用太高（太高了码率也会浪费）。通常，aq-mode=1，aq-strength 给 0.8 比较合理；aq-mode=2，aq-strength 给 0.9 左右，aq-mode=3，aq-strength 给 0.7 左右。

--psy-rd 2.0。psy 是目前 x265 调节锐利度和细节保留的重要工具，低了会糊高了会出现动态瑕疵。默认的 2.0 其实是个不错的数值。如果中低码率编码，可以考虑降低到 1.5左右。

--psy-rdoq 1.0。作用类似 x264 种的 psy-trellis，开一点有助于保留细节和噪点。

--rdoq-level 2。注意默认的--preset medium 下它是 0，这时候 rdoq 是没有用的。slow 及以上自动开启。

--pbratio 1.2，降低 p 帧和 b 帧间画质差距。动漫编码 b 帧数量庞大，且 pb 之间分工不明显，因此降低这个参数对全局有利。

--cbqpoffs -2 --crqpoffs -2，类似 x264 里 chroma-qp-offset。x264 里，开启 psy 同时会降低这两个参数，因为 psy 作用于 luma 平面，会倾向于将码率较多的分配给 luma 平面，所以 x264 会根据 psy 强度自动调整chroma-qp-offset，强行提高 chroma 平面的码率分配。x265 中无此机制，导致 x265 经常出现 chroma 欠码导致的色彩纹理削弱。给-2 左右的 offset 可以较好的缓解问题。

--qcomp 0.65，略高于默认的 0.6，对时域分配采取略保守的策略，来针对中高画质优化。

## 5. VCB-Studio 常用参数分析

vcb-s 常用的参数如下：

```shell
x265-10b --y4m -D 10 --preset slower --deblock -1:-1 --ctu 32 --qg-size 8 --crf 15.0 --pbratio 1.2
--cbqpoffs -2 --crqpoffs -2 --no-sao --me 3 --subme 5 --merange 38 --b-intra --limit-tu 4 --no-amp --ref
4 --weightb --keyint 360 --min-keyint 1 --bframes 6 --aq-mode 1 --aq-strength 0.8 --rd 5 --psy-rd 2.0
--psy-rdoq 1.0 --rdoq-level 2 --no-open-gop --rc-lookahead 80 --scenecut 40 --qcomp 0.65
--no-strong-intra-smoothing --output "EP01.hevc" -
```

编码规范方面，主要是关闭了 open-gop；

tradeoff 方面，主要是基于 preset slower, 然后针对 1080p 的分辨率微淘了--ctu, ME 相关的一些（x265 默认参数针对的是>1080p 分辨率），而动漫特性编码，--ref/bframes 这些则予以加强。amp 实测没有发现在 1080p 下有太多效果，不建议开启。

rc 方面详细调整的较多，主要方向是通过关闭 sao，调节 psy/aq/cqpoff 等来强化细节保留和编码效率。总体来说，aq 针对默认略低，而 psy 相比较默认提高（其实是 rdoq 开启），offset 则手动加强 Chroma 画质（当然，如果你编码 444 请记得手动降低）

## 6. x265 官方的--tune animation

x265 官方给出的--tune animation 多少有点照抄 x264 的意思：

--psy-rd 0.4。从 2.0 直接降到 0.4 也是够拼的；在官方看来 animation 大概就是 flash 动画的效果。不过动漫压制确实可以调低一些，有时候不强调平面细节，希望把码率加到线条上，给到 1.5 左右也是可以的。

--aq-strength 0.4。同上。

--deblock 1:1。同上。

--bframes 增加 2。默认的 bframes 为 4，增加 2 就是 6。这个改动也是比较合理的。