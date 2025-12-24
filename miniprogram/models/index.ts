export type UserRole = 'admin' | 'member';

export interface User {
  id: string;          // OpenID
  nickName: string;
  avatarUrl: string;
  familyId: string | null;
  role: UserRole;
}

export interface Family {
  id: string;
  name: string;
  ownerId: string;
  createTime: number;
}

export interface Recipe {
  id: string;
  familyId: string;
  name: string;
  images: string[];
  content: string;
  creatorId?: string;
  creatorName?: string;
  creatorAvatar?: string;
  createTime: number;
  updateTime: number;
  sortOrder: number;
}

export interface Message {
  id: string;
  familyId: string;
  userId: string;
  userName: string;
  userAvatar: string;
  content: string;
  createTime: number;
  parentId: string | null;
  replyToUserId?: string;
  replyToUserName?: string;
  replies?: Message[];
}

export interface Menu {
  id: string;
  familyId: string;
  date: string;
  mealType: 1 | 2 | 3;
  recipeId: string;
  recipeName: string;
  recipeImg: string;
  userId: string;
  userAvatar: string;
  remark: string;
  createTime: number;
}

// 强制生成 JS 文件，避免 import 报错

