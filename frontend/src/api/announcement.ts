import {request} from '../utils/request';
import type{Announcement,AnnouncementScope,ScopeOptions}from '../types';

export const listAnnouncements=()=>request<Announcement[]>('/announcements');

export const createAnnouncement=(data:{title:string;content:string;category:string;top:boolean;scope:AnnouncementScope;building?:string;unit?:string})=>
  request<Announcement>('/announcements',{method:'POST',body:JSON.stringify(data)});

export const detailAnnouncement=(id:number)=>request<Announcement>(`/announcements/${id}`);

/** 住户确认紧急公告已读（幂等，重复确认不增加人数）。 */
export const confirmAnnouncement=(id:number)=>request<Announcement>(`/announcements/${id}/confirm`,{method:'POST'});

/** 物业发布时可选的楼栋 / 单元（取值来自住户已绑定房产）。 */
export const announcementScopeOptions=(building='')=>request<ScopeOptions>(`/announcement-scope-options${building?`?building=${encodeURIComponent(building)}`:''}`);
