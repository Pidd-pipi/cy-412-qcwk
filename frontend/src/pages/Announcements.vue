<template>
  <section>
    <header class="page-head">
      <div><p class="eyebrow">COMMUNITY</p><h2>社区公告</h2></div>
      <el-button v-if="canPublish" type="primary" @click="dialog=true">发布公告</el-button>
    </header>
    <AnnouncementCard v-for="v in items" :key="v.id" :announcement="v" @open="open"/>
    <EmptyState v-if="!items.length"/>

    <el-dialog v-model="dialog" title="发布公告">
      <el-form label-position="top">
        <el-form-item label="标题"><el-input v-model="form.title" placeholder="标题"/></el-form-item>
        <el-form-item label="内容"><el-input v-model="form.content" type="textarea" placeholder="内容"/></el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category">
            <el-option label="通知" value="通知"/>
            <el-option label="活动" value="活动"/>
            <el-option label="紧急" value="紧急"/>
          </el-select>
        </el-form-item>
        <el-form-item label="通知范围">
          <el-radio-group v-model="form.scope">
            <el-radio value="all">全小区</el-radio>
            <el-radio value="building">按楼栋</el-radio>
            <el-radio value="unit">按单元</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.scope!=='all'" label="楼栋 / 单元">
          <div class="property">
            <el-input v-model="form.scope_building" placeholder="楼栋，如 1栋"/>
            <el-input v-if="form.scope==='unit'" v-model="form.scope_unit" placeholder="单元，如 2单元"/>
          </div>
        </el-form-item>
        <el-form-item><el-checkbox v-model="form.top">置顶</el-checkbox></el-form-item>
      </el-form>
      <template #footer><el-button type="primary" @click="publish">发布</el-button></template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detail?.title" width="560px">
      <el-space wrap>
        <el-tag size="small">{{detail?.category}}</el-tag>
        <el-tag size="small" type="warning">{{detailScopeText}}</el-tag>
      </el-space>
      <p class="detail-content">{{detail?.content}}</p>
      <template v-if="isStaff">
        <template v-if="detail?.category==='紧急'">
          <small>已确认 {{detail?.confirmed_count}} 人 · 未确认 {{detail?.unconfirmed_users?.length||0}} 人</small>
          <div class="confirm-list">
            <el-tag v-for="u in detail?.unconfirmed_users" :key="u.id" type="info" size="small">
              {{u.nickname}}（{{[u.building,u.unit,u.room].filter(Boolean).join(' ')||u.phone}}）
            </el-tag>
            <small v-if="!detail?.unconfirmed_users?.length" class="muted">范围内住户均已确认</small>
          </div>
        </template>
        <small v-else>已读 {{detail?.read_count}}</small>
      </template>
      <template v-else>
        <small v-if="detail?.category!=='紧急'">已读 {{detail?.read_count}}</small>
        <div v-else class="resident-confirm">
          <el-tag v-if="detail?.confirmed" type="success">已确认，感谢配合</el-tag>
          <template v-else>
            <el-tag type="danger">紧急通知，请确认已读</el-tag>
            <el-button type="primary" size="small" :loading="confirming" @click="confirm">我已确认</el-button>
          </template>
        </div>
      </template>
    </el-dialog>
  </section>
</template>
<script setup lang="ts">
import{ref,computed,onMounted}from'vue';
import{useRoute}from'vue-router';
import{ElMessage}from'element-plus';
import{listAnnouncements,createAnnouncement,detailAnnouncement,confirmAnnouncement}from'../api/announcement';
import{authStore}from'../stores/authStore';
import type{Announcement,AnnouncementScope}from'../types';
import AnnouncementCard from'../components/common/AnnouncementCard.vue';
import EmptyState from'../components/common/EmptyState.vue';

const items=ref<Announcement[]>([]),
  dialog=ref(false),detailVisible=ref(false),detail=ref<Announcement>(),
  confirming=ref(false),
  form=ref<{title:string;content:string;category:string;scope:AnnouncementScope;scope_building:string;scope_unit:string;top:boolean}>({title:'',content:'',category:'通知',scope:'all',scope_building:'',scope_unit:'',top:false});
const route=useRoute();
const isStaff=computed(()=>['staff','admin'].includes(authStore.user?.role||''));
const canPublish=computed(()=>isStaff.value);
const detailScopeText=computed(()=>{
  const a=detail.value;if(!a)return'';
  if(a.scope==='building')return`楼栋 · ${a.scope_building}`;
  if(a.scope==='unit')return`单元 · ${a.scope_building} ${a.scope_unit}`;
  return'全小区';
});

async function load(){items.value=await listAnnouncements()}
async function open(id:number){
  detail.value=await detailAnnouncement(id);
  detailVisible.value=true;
  load();
}
async function publish(){
  if(form.value.scope!=='all'&&!form.value.scope_building.trim()){ElMessage.warning('请填写通知范围对应的楼栋');return}
  if(form.value.scope==='unit'&&!form.value.scope_unit.trim()){ElMessage.warning('请填写单元');return}
  try{
    await createAnnouncement({...form.value});
    dialog.value=false;
    form.value={title:'',content:'',category:'通知',scope:'all',scope_building:'',scope_unit:'',top:false};
    ElMessage.success('公告已发布');
    load();
  }catch(e){ElMessage.error((e as Error).message)}
}
async function confirm(){
  if(!detail.value)return;
  confirming.value=true;
  try{
    detail.value=await confirmAnnouncement(detail.value.id);
    ElMessage.success('已确认');
    load();
  }catch(e){ElMessage.error((e as Error).message)}finally{confirming.value=false}
}

onMounted(async()=>{
  await load();
  const qid=route.query.open;
  if(typeof qid==='string'&&items.value.some(v=>v.id===Number(qid)))open(Number(qid));
});
</script>
