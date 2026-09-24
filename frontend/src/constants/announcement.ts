import type {AnnouncementScope} from '../types';

// 公告范围，与后端 internal/constants/announcement.go 对应。
export const ANNOUNCEMENT_SCOPE:Record<Uppercase<AnnouncementScope>,AnnouncementScope>={ALL:'all',BUILDING:'building',UNIT:'unit'};

export const SCOPE_TEXT:Record<AnnouncementScope,string>={all:'全小区',building:'指定楼栋',unit:'指定单元'};

/** 紧急公告分类，需住户确认已读。 */
export const ANNOUNCEMENT_CATEGORY_URGENT='紧急';

/** 公告范围的可读描述：全小区 / 3栋 / 1栋2单元 */
export function scopeText(a:{scope:AnnouncementScope;building?:string;unit?:string}):string{
  if(a.scope==='building') return a.building||'指定楼栋';
  if(a.scope==='unit') return `${a.building||''}${a.unit||''}`||'指定单元';
  return '全小区';
}
