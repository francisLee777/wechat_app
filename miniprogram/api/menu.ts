import { request } from '../utils/request';
import type { Menu } from '../models/index';

export interface MenuListResponse {
  breakfast: Menu[];
  lunch: Menu[];
  dinner: Menu[];
}

export async function getMenuList(date: string): Promise<MenuListResponse> {
  const res = await request<any>(`/menu/list?date=${date}`, 'GET');
  // 转换 ID 为字符串，防止精度丢失
  const convert = (list: any[]) => list.map(m => ({
    ...m,
    id: String(m.id),
    familyId: String(m.familyId),
    recipeId: String(m.recipeId)
  }));
  
  return {
    breakfast: convert(res.breakfast || []),
    lunch: convert(res.lunch || []),
    dinner: convert(res.dinner || [])
  };
}

export async function addMenu(date: string, mealType: number, recipeId: string, remark: string = ''): Promise<Menu> {
  const res = await request<any>('/menu/add', 'POST', {
    date,
    mealType,
    recipeId: parseInt(recipeId),
    remark
  });
  return {
    ...res,
    id: String(res.id),
    familyId: String(res.familyId),
    recipeId: String(res.recipeId)
  };
}

export async function deleteMenu(id: string): Promise<void> {
  await request(`/menu/delete?id=${id}`, 'POST');
}
