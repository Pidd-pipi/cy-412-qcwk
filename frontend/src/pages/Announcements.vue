<template>
  <section>
    <header class="page-head">
      <div>
        <p class="eyebrow">COMMUNITY</p>
        <h2>社区公告</h2>
      </div>
      <el-button v-if="canPublish" type="primary" @click="openPublish">发布公告</el-button>
    </header>

    <AnnouncementCard v-for="v in items" :key="v.id" :announcement="v" @open="open" @confirm="confirm" />
    <EmptyState v-if="!items.length" />

    <!-- 发布公告 -->
    <el-dialog v-model="dialog" title="发布公告" width="520px">
      <el-form label-position="top">
        <el-form-item label="标题" required>
          <el-input v-model="form.title" placeholder="标题" />
        </el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="form.content" type="textarea" :rows="4" placeholder="内容" />
        </el-form-item>
        <el-form-item label="分类" required>
          <el-select v-model="form.category">
            <el-option label="通知" value="通知" />
            <el-option label="活动" value="活动" />
            <el-option label="紧急" value="紧急" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布范围" required>
          <el-radio-group v-model="form.scope" @change="onScopeChange">
            <el-radio value="all">全小区</el-radio>
            <el-radio value="building">指定楼栋</el-radio>
            <el-radio value="unit">指定单元</el-radio>
          </el-radio-group>
        </el-form-item>
        <div v-if="form.scope!=='all'" class="scope-row">
          <el-select v-model="form.building" placeholder="选择楼栋" @change="onBuildingChange">
            <el-option v-for="b in scopeBuildings" :key="b" :label="b" :value="b" />
          </el-select>
          <el-select v-if="form.scope==='unit'" v-model="form.unit" placeholder="选择单元">
            <el-option v-for="u in scopeUnits" :key="u" :label="u" :value="u" />
          </el-select>
        </div>
        <el-alert v-if="form.category==='紧急'" type="error" :closable="false" show-icon title="紧急公告发布后，范围内住户需点击“确认已读”，物业可查看确认人数与未确认名单。" />
        <el-form-item style="margin-top:12px">
          <el-checkbox v-model="form.top">置顶</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" @click="publish">发布</el-button>
      </template>
    </el-dialog>

    <!-- 公告详情 -->
    <el-dialog v-model="detailVisible" :title="detail?.title" width="560px">
      <div class="detail-meta">
        <el-tag size="small">{{detail?.category}}</el-tag>
        <el-tag size="small" type="info" effect="plain" style="margin-left:6px">{{detail?scopeText(detail):''}}</el-tag>
        <span style="margin-left:8px">{{detail?.publisher?.nickname}} · {{detail?.publish_at?.replace('T',' ').slice(0,16)}}</span>
      </div>
      <p style="line-height:1.7">{{detail?.content}}</p>

      <!-- 普通公告：阅读人数 -->
      <small v-if="detail&&detail.category!=='紧急'" class="detail-meta">已读 {{detail.read_count}}</small>

      <template v-else-if="detail">
        <!-- 住户视角 -->
        <div v-if="isResident">
          <el-button v-if="!detail.confirmed" type="danger" @click="confirm(detail.id)">确认已读</el-button>
          <el-tag v-else type="success">你已确认</el-tag>
        </div>
        <!-- 物业视角：确认人数 + 未确认名单 -->
        <div v-else>
          <p class="detail-meta">
            已确认 <b>{{detail.confirmed_count||0}}</b> 人，
            未确认 <b style="color:#e6a23c">{{detail.unconfirmed_users?.length||0}}</b> 人
          </p>
          <div v-if="detail.unconfirmed_users?.length" class="unconfirmed-box">
            <p v-for="u in detail.unconfirmed_users" :key="u.id">{{u.nickname}}（{{u.phone}} {{u.building}}{{u.unit}}{{u.room}}）</p>
          </div>
          <el-tag v-else type="success">范围内住户已全部确认</el-tag>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue';
import {useRoute} from 'vue-router';
import {ElMessage} from 'element-plus';
import {listAnnouncements, createAnnouncement, detailAnnouncement, confirmAnnouncement, announcementScopeOptions} from '../api/announcement';
import {authStore} from '../stores/authStore';
import type {Announcement, AnnouncementScope} from '../types';
import {scopeText} from '../constants/announcement';
import AnnouncementCard from '../components/common/AnnouncementCard.vue';
import EmptyState from '../components/common/EmptyState.vue';

const route = useRoute();
const items = ref<Announcement[]>([]);
const dialog = ref(false);
const detailVisible = ref(false);
const detail = ref<Announcement>();
const scopeBuildings = ref<string[]>([]);
const scopeUnits = ref<string[]>([]);
const form = ref({title:'', content:'', category:'通知', top:false, scope:'all' as AnnouncementScope, building:'', unit:''});

const canPublish = computed(()=>['staff','admin'].includes(authStore.user?.role||''));
const isResident = computed(()=>authStore.user?.role==='resident');

async function load() {
  items.value = await listAnnouncements();
}

// 从工作台跳转携带 open 参数时自动打开详情（仅一次，随后清理参数避免重复打开）。
async function openFromQuery() {
  const openId = Number(route.query.open);
  if (openId) {
    await open(openId);
    history.replaceState(null, '', window.location.pathname);
  }
}

function openPublish() {
  form.value = {title:'', content:'', category:'通知', top:false, scope:'all', building:'', unit:''};
  scopeUnits.value = [];
  dialog.value = true;
  loadScopeBuildings();
}

async function loadScopeBuildings() {
  const opts = await announcementScopeOptions();
  scopeBuildings.value = opts.buildings;
}

async function onScopeChange(s: AnnouncementScope) {
  form.value.unit = '';
  scopeUnits.value = [];
  if (s !== 'all') {
    await loadScopeBuildings();
    if (form.value.building) await onBuildingChange(form.value.building);
  }
}

async function onBuildingChange(b: string) {
  form.value.unit = '';
  if (!b) { scopeUnits.value = []; return; }
  const opts = await announcementScopeOptions(b);
  scopeUnits.value = opts.units;
}

async function publish() {
  if (!form.value.title || !form.value.content) {
    ElMessage.warning('请填写标题和内容');
    return;
  }
  if (form.value.scope==='building' && !form.value.building) {
    ElMessage.warning('请选择楼栋');
    return;
  }
  if (form.value.scope==='unit' && (!form.value.building || !form.value.unit)) {
    ElMessage.warning('请选择楼栋和单元');
    return;
  }
  try {
    await createAnnouncement({...form.value});
    dialog.value = false;
    ElMessage.success('公告已发布');
    load();
  } catch (e) {
    ElMessage.error((e as Error).message);
  }
}

async function open(id: number) {
  detail.value = await detailAnnouncement(id);
  detailVisible.value = true;
  load();
}

async function confirm(id: number) {
  try {
    await confirmAnnouncement(id);
    ElMessage.success('已确认');
    if (detailVisible.value && detail.value?.id===id) detail.value = await detailAnnouncement(id);
    load();
  } catch (e) {
    ElMessage.error((e as Error).message);
  }
}

onMounted(async()=>{
  await load();
  await openFromQuery();
});
</script>
