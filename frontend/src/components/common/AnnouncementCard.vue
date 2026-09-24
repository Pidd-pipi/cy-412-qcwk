<template>
  <article class="card announcement" @click="$emit('open',announcement.id)">
    <div>
      <el-tag v-if="announcement.top" type="danger" size="small">置顶</el-tag>
      <el-tag :type="urgent?'danger':''" size="small">{{announcement.category}}</el-tag>
      <el-tag type="info" size="small" effect="plain">{{scopeText(announcement)}}</el-tag>
      <h3>{{announcement.title}}</h3>
      <p>{{announcement.content}}</p>
    </div>
    <footer class="announcement-foot">
      <!-- 普通公告：继续展示现有阅读人数 -->
      <small v-if="!urgent">已读 {{announcement.read_count}}</small>
      <!-- 紧急公告·住户：确认按钮或已确认状态 -->
      <template v-else-if="isResident">
        <el-button v-if="!announcement.confirmed" type="danger" size="small" @click.stop="$emit('confirm',announcement.id)">确认已读</el-button>
        <el-tag v-else type="success" size="small">已确认</el-tag>
      </template>
      <!-- 紧急公告·物业：确认人数与未确认名单 -->
      <div v-else class="urgent-stats" @click.stop>
        <small>已确认 {{announcement.confirmed_count||0}} / 共 {{(announcement.confirmed_count||0)+(announcement.unconfirmed_users?.length||0)}} 人</small>
        <el-popover v-if="announcement.unconfirmed_users?.length" placement="bottom" :width="220" trigger="hover">
          <template #reference><el-link type="warning" :underline="false" class="unconfirmed-link">未确认 {{announcement.unconfirmed_users.length}} 人</el-link></template>
          <p class="unconfirmed-title">未确认名单</p>
          <p v-for="u in announcement.unconfirmed_users" :key="u.id" class="unconfirmed-item">{{u.nickname}}（{{u.building}}{{u.unit}}{{u.room}}）</p>
        </el-popover>
        <el-tag v-else type="success" size="small">全部已确认</el-tag>
      </div>
    </footer>
  </article>
</template>
<script setup lang="ts">
import{computed}from'vue';
import type{Announcement}from'../../types';
import{ANNOUNCEMENT_CATEGORY_URGENT,scopeText}from'../../constants/announcement';
import{authStore}from'../../stores/authStore';

const props=defineProps<{announcement:Announcement}>();
defineEmits<{open:[id:number];confirm:[id:number]}>();
const urgent=computed(()=>props.announcement.category===ANNOUNCEMENT_CATEGORY_URGENT);
const isResident=computed(()=>authStore.user?.role==='resident');
</script>
