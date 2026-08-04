<template>
  <Teleport to="body">
    <div v-if="visible" class="cfg-overlay" @click.self="cancel">
      <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:110;width:640px;max-height:85vh;overflow-y:auto;padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
          <h3 style="font-size:18px;font-weight:700;">新建编码配置</h3>
          <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="cancel">取消</button>
        </div>

        <div v-if="configStep === 0" style="text-align:center;padding:20px 0;">
          <div style="font-size:15px;font-weight:600;margin-bottom:20px;">选择配置模式</div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
            <div class="hover-span" style="padding:24px;cursor:pointer;text-align:center;border-radius:12px;background:rgba(255,255,255,0.02);border:1px solid rgba(255,255,255,0.06);" @click="startSimple">
              <div style="font-size:32px;margin-bottom:8px;">⚡</div>
              <div style="font-weight:600;margin-bottom:4px;">简易模式</div>
              <div style="font-size:12px;color:#8A8F98;">基础参数，快速配置</div>
            </div>
            <div class="hover-span" style="padding:24px;cursor:pointer;text-align:center;border-radius:12px;background:rgba(255,255,255,0.02);border:1px solid rgba(255,255,255,0.06);" @click="startPro">
              <div style="font-size:32px;margin-bottom:8px;">🔧</div>
              <div style="font-weight:600;margin-bottom:4px;">专业模式</div>
              <div style="font-size:12px;color:#8A8F98;">完整参数，精细调优</div>
            </div>
          </div>
        </div>

        <div v-if="configStep === 1 && configMode === 'simple'" style="display:flex;flex-direction:column;gap:16px;">
          <div style="font-weight:600;">简易配置</div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">视频编码器</div>
            <select v-model="newConfig.encoder">
              <option value="x264">x264 (H.264)</option>
              <option value="x265">x265 (H.265)</option>
              <option value="h264_nvenc">NVENC H.264</option>
              <option value="hevc_nvenc">NVENC H.265</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">音频编码器</div>
            <select v-model="newConfig.audioEncoder">
              <option value="flac">FLAC (无损)</option>
              <option value="aac">AAC (有损)</option>
              <option value="opus">Opus (有损)</option>
            </select>
          </div>
          <div v-if="newConfig.audioEncoder !== 'flac'">
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">音频码率</div>
            <select v-model="newConfig.audioBitrate">
              <option value="128">128 kbps</option>
              <option value="192">192 kbps</option>
              <option value="256">256 kbps</option>
              <option value="320">320 kbps</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">视频质量</div>
            <select v-model="newConfig.quality">
              <option value="lossless">无损</option>
              <option value="high">高</option>
              <option value="medium">中</option>
              <option value="low">低</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">编码速度</div>
            <select v-model="newConfig.speed">
              <option value="slow">慢 (高质量)</option>
              <option value="medium">平衡</option>
              <option value="fast">快 (低质量)</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">位深</div>
            <select v-model="newConfig.depth">
              <option value="10">10-bit</option>
              <option value="8">8-bit</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">配置名称</div>
            <input type="text" v-model="newConfig.name" placeholder="输入配置名称" />
          </div>
          <div style="display:flex;justify-content:flex-end;gap:8px;margin-top:8px;">
            <button class="btn-ghost" @click="cancel">取消</button>
            <button class="btn-primary" @click="saveNewConfig">保存配置</button>
          </div>
        </div>

        <div v-if="configStep === 1 && configMode === 'pro'">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
            <h3 style="font-weight:600;">专业配置 — 视频参数</h3>
            <div style="display:flex;align-items:center;gap:8px;">
              <select v-model="newConfig.encoder" style="width:180px;height:40px;">
                <option value="x264">x264 (H.264)</option>
                <option value="x265">x265 (H.265)</option>
                <option value="h264_nvenc">NVIDIA NVENC H.264</option>
                <option value="hevc_nvenc">NVIDIA NVENC H.265</option>
                <option value="h264_qsv">Intel QSV H.264</option>
                <option value="hevc_qsv">Intel QSV H.265</option>
                <option value="h264_amf">AMD AMF H.264</option>
                <option value="hevc_amf">AMD AMF H.265</option>
              </select>
              <button class="btn-ghost" @click="openHelp" style="height:40px;padding:0 10px;display:flex;align-items:center;gap:4px;font-size:12px;">ⓘ 帮助</button>
            </div>
          </div>

          <template v-if="newConfig.encoder === 'x264' || newConfig.encoder === 'x265'">
            <details v-for="group in cpuParamGroups" :key="group.name" class="glass" style="padding:12px 16px;margin-bottom:8px;" open>
              <summary style="font-weight:500;cursor:pointer;font-size:14px;">{{ group.name }}</summary>
              <div style="margin-top:12px;">
                <div v-for="param in group.params" :key="param.key" style="margin-bottom:8px;">
                  <div style="display:flex;align-items:center;justify-content:space-between;">
                    <span style="font-size:13px;color:#8A8F98;">{{ param.label }}</span>
                    <input v-if="param.type === 'number'" type="number" v-model="newConfig.params[param.key]" style="width:120px;text-align:right;" />
                    <select v-else v-model="newConfig.params[param.key]" style="width:160px;">
                      <option v-for="o in param.options" :key="o.value" :value="o.value">{{ o.label }}</option>
                    </select>
                  </div>
                  <div v-if="param.hint" style="font-size:11px;color:rgba(255,255,255,0.3);text-align:right;margin-top:2px;">
                    推荐值: {{ param.hint }}
                  </div>
                </div>
              </div>
            </details>
          </template>

          <template v-else>
            <div class="glass" style="padding:16px;">
              <div v-for="p in gpuParams" style="margin-bottom:12px;">
                <div style="display:flex;align-items:center;justify-content:space-between;">
                  <span style="font-size:13px;color:#8A8F98;">{{ p.label }}</span>
                  <select v-model="newConfig.params[p.key]" style="width:160px;">
                    <option v-for="o in p.options" :key="o.value" :value="o.value">{{ o.label }}</option>
                  </select>
                </div>
              </div>
            </div>
          </template>

          <button class="btn-primary" style="margin-top:12px;" @click="configStep = 2">下一步: 音频配置</button>
        </div>

        <div v-if="configStep === 2 && configMode === 'pro'" style="display:flex;flex-direction:column;gap:16px;">
          <div style="font-weight:600;">专业配置 — 音频参数</div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">音频编码器</div>
            <select v-model="newConfig.audioEncoder">
              <option value="flac">FLAC (无损)</option>
              <option value="aac">AAC (有损)</option>
              <option value="opus">Opus</option>
              <option value="copy">复制原始音轨</option>
            </select>
          </div>
          <div v-if="newConfig.audioEncoder !== 'flac' && newConfig.audioEncoder !== 'copy'">
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">音频码率</div>
            <select v-model="newConfig.audioBitrate">
              <option value="128">128 kbps</option>
              <option value="192">192 kbps</option>
              <option value="256">256 kbps</option>
              <option value="320">320 kbps</option>
            </select>
          </div>
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">配置名称</div>
            <input type="text" v-model="newConfig.name" placeholder="输入配置名称" />
          </div>
          <div style="display:flex;justify-content:space-between;margin-top:8px;">
            <button class="btn-ghost" @click="configStep = 1">上一步</button>
            <div style="display:flex;gap:8px;">
              <button class="btn-ghost" @click="cancel">取消</button>
              <button class="btn-primary" @click="saveNewConfig">保存配置</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showHelp" style="position:fixed;inset:0;z-index:120;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);" @click.self="showHelp = false">
      <div class="glass" style="width:700px;max-height:80vh;display:flex;flex-direction:column;">
        <div style="display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.06);">
          <div style="display:flex;align-items:center;gap:12px;">
            <h3 style="font-weight:600;font-size:16px;">参数帮助</h3>
            <select v-model="helpDoc" style="width:auto;font-size:13px;">
              <option value="x264">x264</option>
              <option value="x265">x265</option>
              <option value="nvenc">NVENC</option>
              <option value="qsv">Intel QSV</option>
            </select>
          </div>
          <button @click="showHelp = false" style="color:#8A8F98;font-size:20px;cursor:pointer;background:none;border:none;">✕</button>
        </div>
        <div style="flex:1;overflow-y:auto;padding:20px;font-size:14px;line-height:1.8;color:#EDEDEF;">
          <template v-if="helpDoc === 'x264'">
            <h2 style="color:#5E6AD2;font-weight:600;margin-bottom:12px;">x264 编码参数</h2>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">基础设置</h3>
            <p><strong>crf</strong> — 固定质量模式 (0-51)。动漫蓝光 BDRip 建议 16-18。值越低质量越高。</p>
            <p><strong>preset</strong> — 速度/质量平衡。推荐 slow、slower 或 veryslow。</p>
            <p><strong>tune</strong> — 片源优化: film (电影), animation (动漫), grain (噪点), stillimage (静态)。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">帧类型</h3>
            <p><strong>keyint</strong> — GOP 区间。24fps 推荐 600。</p>
            <p><strong>bframes</strong> — B 帧数。动漫 8-12，电影 4-8。</p>
            <p><strong>ref</strong> — 参考帧。推荐 6-13。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">码率控制</h3>
            <p><strong>qcomp</strong> — 时域分配灵活度。开 mbtree 推荐 0.7-0.8。</p>
            <p><strong>aq-mode</strong> — 自适应量化。mode=3 最适合动漫。</p>
            <p><strong>aq-strength</strong> — AQ 强度。动漫 0.6-1.0。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">运动估计</h3>
            <p><strong>me</strong> — 运动搜索算法。推荐 umh 或 tesa。</p>
            <p><strong>subme</strong> — 亚像素优化。推荐 10。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">心理视觉</h3>
            <p><strong>psy-rd</strong> — 纹理锐度保留。动漫 0.4-1.0。</p>
            <p><strong>psy-trellis</strong> — 细节保留。开 mbtree 推荐 0.1-0.15。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">其他</h3>
            <p><strong>deblock</strong> — 去色块。高画质建议 -1:-1。</p>
          </template>
          <template v-else-if="helpDoc === 'x265'">
            <h2 style="color:#5E6AD2;font-weight:600;margin-bottom:12px;">x265 编码参数</h2>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">基础设置</h3>
            <p><strong>crf</strong> — 高质量编码推荐 15-18。</p>
            <p><strong>preset</strong> — 推荐 slow 或 slower。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">分块</h3>
            <p><strong>ctu</strong> — 1080p 限制为 32。</p>
            <p><strong>qg-size</strong> — 推荐 8。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">码率控制</h3>
            <p><strong>qcomp</strong> — 推荐 0.65。</p>
            <p><strong>no-sao</strong> — 高画质建议关闭 (涂抹效果严重)。</p>
            <p><strong>aq-mode</strong> — mode=1 高画质，mode=2 高效率。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">运动</h3>
            <p><strong>me</strong> — 推荐 star (me=3)。</p>
            <p><strong>subme</strong> — 推荐 5。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">色彩</h3>
            <p><strong>cbqpoffs/crqpoffs</strong> — 推荐 -2。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">其他</h3>
            <p><strong>deblock</strong> — 高画质建议 -1:-1。</p>
            <p><strong>keyint</strong> — 推荐 360。</p>
          </template>
          <template v-else-if="helpDoc === 'nvenc'">
            <h2 style="color:#5E6AD2;font-weight:600;margin-bottom:12px;">NVIDIA NVENC 编码参数</h2>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">基础设置</h3>
            <p><strong>preset</strong> — p1 (最快) 到 p7 (最慢/最高质量)。推荐 p7。</p>
            <p><strong>tune</strong> — hq (高质量), ll (低延迟), ull (超低延迟)。</p>
            <p><strong>rc</strong> — 码率控制: cbr (恒定), vbr (可变), constqp (恒定 QP)。推荐 vbr。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">高级设置</h3>
            <p><strong>multipass</strong> — qres (1/4 分辨率), fullres (全分辨率)。推荐 qres。</p>
            <p><strong>b_ref_mode</strong> — each (独立参考), middle (中间参考)。推荐 middle。</p>
            <p><strong>lookahead</strong> — 前瞻帧数。推荐 32。</p>
            <p><strong>bframes</strong> — 最大 B 帧数。推荐 3-5。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">质量</h3>
            <p><strong>aq-strength</strong> — 自适应量化强度 (1-15)。推荐 8。</p>
          </template>
          <template v-else-if="helpDoc === 'qsv'">
            <h2 style="color:#5E6AD2;font-weight:600;margin-bottom:12px;">Intel QSV 编码参数</h2>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">基础设置</h3>
            <p><strong>preset</strong> — veryfast, faster, fast, medium, slow, slower, veryslow。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">码率控制</h3>
            <p><strong>rc</strong> — CQP (恒定 QP), VBR, CBR, ICQ (智能)。</p>
            <h3 style="color:#94A3B8;margin:16px 0 8px;">高级</h3>
            <p><strong>bframes</strong> — 最大 B 帧数。推荐 4。</p>
            <p><strong>refs</strong> — 参考帧数。推荐 4。</p>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { api } from '@/api'

const props = defineProps<{ visible: boolean; editConfig?: any }>()
const emit = defineEmits<{ close: []; saved: [config: any]; error: [msg: string] }>()

const configStep = ref(0)
const configMode = ref<'simple' | 'pro'>('simple')

const defaultParams: Record<string, any> = {
  crf: 18, preset: 'medium', tune: 'animation', keyint: 600, bframes: 8, ref: 13,
  qcomp: 0.75, 'aq-mode': '3', 'aq-strength': 0.8, me: 'star', subme: 10, merange: 24,
}

const newConfig = reactive({
  name: '', encoder: 'x265',
  audioEncoder: 'flac', audioBitrate: '192',
  quality: 'high', speed: 'medium', depth: '10',
  params: { ...defaultParams },
})

const cpuParamGroups = [
  {
    name: '基础',
    params: [
      { key: 'crf', label: 'CRF (质量)', type: 'number', hint: '15 — 18' },
      { key: 'preset', label: '预设', type: 'select', options: [
        { value: 'ultrafast', label: 'ultrafast' }, { value: 'veryfast', label: 'veryfast' },
        { value: 'faster', label: 'faster' }, { value: 'fast', label: 'fast' },
        { value: 'medium', label: 'medium' }, { value: 'slow', label: 'slow (推荐)' },
        { value: 'slower', label: 'slower (推荐)' }, { value: 'veryslow', label: 'veryslow (推荐)' },
      ] },
      { key: 'tune', label: '优化', type: 'select', options: [
        { value: 'animation', label: 'animation' }, { value: 'film', label: 'film' },
        { value: 'grain', label: 'grain' }, { value: 'stillimage', label: 'stillimage' },
      ] },
    ]
  },
  {
    name: '帧类型',
    params: [
      { key: 'keyint', label: 'GOP 区间', type: 'number', hint: '600' },
      { key: 'bframes', label: 'B 帧数', type: 'number', hint: '8' },
      { key: 'ref', label: '参考帧', type: 'number', hint: '13' },
    ]
  },
  {
    name: '码率控制',
    params: [
      { key: 'qcomp', label: 'QComp', type: 'number', hint: '0.75' },
      { key: 'aq-mode', label: 'AQ 模式', type: 'select', options: [
        { value: '1', label: '1 - 标准' }, { value: '2', label: '2 - 自动' }, { value: '3', label: '3 - 动漫推荐' },
      ] },
      { key: 'aq-strength', label: 'AQ 强度', type: 'number', hint: '0.8' },
    ]
  },
  {
    name: '运动估计',
    params: [
      { key: 'me', label: '运动搜索', type: 'select', options: [
        { value: 'hex', label: 'hex' }, { value: 'umh', label: 'umh' },
        { value: 'star', label: 'star' }, { value: 'tesa', label: 'tesa' },
      ] },
      { key: 'subme', label: '亚像素', type: 'number', hint: '10' },
      { key: 'merange', label: '搜索范围', type: 'number', hint: '24' },
    ]
  },
]

const gpuParams = [
  { key: 'preset', label: '预设', options: [
    { value: 'p1', label: 'p1 - 最快' }, { value: 'p2', label: 'p2' },
    { value: 'p3', label: 'p3' }, { value: 'p4', label: 'p4 - 中等' },
    { value: 'p5', label: 'p5' }, { value: 'p6', label: 'p6' },
    { value: 'p7', label: 'p7 - 最高质量' },
  ] },
  { key: 'rc', label: '码率控制', options: [
    { value: 'vbr', label: 'VBR - 可变码率' }, { value: 'cbr', label: 'CBR - 恒定码率' },
    { value: 'constqp', label: 'ConstQP - 恒定 QP' },
  ] },
  { key: 'b_ref_mode', label: 'B 帧参考模式', options: [
    { value: 'disabled', label: '禁用' }, { value: 'each', label: '每个' }, { value: 'middle', label: '中间 (推荐)' },
  ] },
  { key: 'multipass', label: '多通道', options: [
    { value: 'disabled', label: '禁用' }, { value: 'qres', label: 'qres (推荐)' }, { value: 'fullres', label: 'fullres' },
  ] },
]

const showHelp = ref(false)
const helpDoc = ref('x264')

function openHelp() {
  if (newConfig.encoder.startsWith('x265')) helpDoc.value = 'x265'
  else if (newConfig.encoder.startsWith('x264')) helpDoc.value = 'x264'
  else if (newConfig.encoder.includes('qsv')) helpDoc.value = 'qsv'
  else helpDoc.value = 'nvenc'
  showHelp.value = true
}

function startSimple() { configMode.value = 'simple'; configStep.value = 1 }
function startPro() { configMode.value = 'pro'; configStep.value = 1 }

function resetForm() {
  configStep.value = 0
  newConfig.name = ''
  newConfig.encoder = 'x265'
  newConfig.audioEncoder = 'flac'
  newConfig.audioBitrate = '192'
  newConfig.quality = 'high'
  newConfig.speed = 'medium'
  newConfig.depth = '10'
  newConfig.params = { ...defaultParams }
}

function cancel() {
  resetForm()
  emit('close')
}

watch(() => props.editConfig, (cfg) => {
  if (cfg) {
    configMode.value = 'pro'
    newConfig.name = cfg.name || ''
    newConfig.encoder = cfg.encoder || cfg.video_encoder || 'x265'
    newConfig.audioEncoder = (cfg.audio?.codec || 'flac').toLowerCase()
    newConfig.audioBitrate = cfg.audio?.bitrate ? String(cfg.audio.bitrate).replace('kbps', '') : '192'
    newConfig.params = { ...defaultParams, ...cfg.params }
    configStep.value = 1
  }
}, { immediate: true })

async function saveNewConfig() {
  const name = newConfig.name || '未命名配置'
  const mode = newConfig.encoder.includes('nvenc') || newConfig.encoder.includes('qsv') || newConfig.encoder.includes('amf') ? 'gpu' : 'cpu'
  let params: Record<string, any>
  if (configMode.value === 'simple') {
    const qualityMap: Record<string, number> = { lossless: 0, high: 15, medium: 20, low: 23 }
    const speedMap: Record<string, string> = { slow: 'slower', medium: 'medium', fast: 'veryfast' }
    params = { crf: qualityMap[newConfig.quality] || 18, preset: speedMap[newConfig.speed] || 'medium' }
  } else {
    params = { ...newConfig.params }
  }
  const audioEncoder = newConfig.audioEncoder
  const audioBitrate = newConfig.audioBitrate
  const audio: any = { codec: audioEncoder.toUpperCase() }
  if (audioEncoder !== 'flac' && audioEncoder !== 'copy') {
    audio.bitrate = audioBitrate + 'kbps'
  }

  try {
    const payload = {
      name,
      encoder_type: mode === 'gpu' ? 'gpu' : 'cpu',
      video_encoder: newConfig.encoder,
      video_params: JSON.stringify(params),
      audio_tracks: JSON.stringify(audio),
      subtitle_tracks: '[]',
      chapters_enabled: true,
      output_muxer: 'mkvmerge',
    }
    if (props.editConfig) {
      await api.configs.update(props.editConfig.id, payload)
    } else {
      await api.configs.create(payload)
    }
    const result: any = { name, encoder: newConfig.encoder, mode, isPreset: false, params, audio }
    if (props.editConfig) {
      result.id = props.editConfig.id
    }
    emit('saved', result)
    resetForm()
  } catch (e: any) {
    emit('error', e.message)
  }
}
</script>

<style scoped>
.cfg-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  z-index: 110;
}
.hover-span:hover { background: rgba(94,106,210,0.1); }
</style>
