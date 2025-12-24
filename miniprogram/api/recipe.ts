import { request } from '../utils/request';
import type { Recipe } from '../models/index';
import { toRelativePath } from '../utils/request';

// Helper to ensure we store/pass relative paths
const normalizeImageUrl = (url: string) => toRelativePath(url);

export async function getRecipes(keyword: string = ''): Promise<Recipe[]> {
  const res = await request<any[]>('/recipe/list', 'GET');
  let recipes = res.map(r => ({
    id: String(r.id),
    familyId: String(r.familyId),
    name: r.name,
    images: Array.isArray(r.images) ? r.images.map(normalizeImageUrl) : [],
    content: r.content,
    creatorId: r.creatorId,
    creatorName: r.creatorName,
    creatorAvatar: normalizeImageUrl(r.creatorAvatar),
    createTime: r.createTime, 
    updateTime: r.updateTime,
    sortOrder: r.sortOrder
  }));

  if (keyword) {
    recipes = recipes.filter(r => r.name.includes(keyword) || r.content.includes(keyword));
  }
  return recipes;
}

export async function getRecipeById(id: string): Promise<Recipe | undefined> {
  try {
    const res = await request<any>(`/recipe/info?id=${id}`, 'GET');
    return {
      id: String(res.id),
      familyId: String(res.familyId),
      name: res.name,
      images: Array.isArray(res.images) ? res.images.map(normalizeImageUrl) : [],
      content: res.content,
      creatorId: res.creatorId,
      creatorName: res.creatorName,
      creatorAvatar: normalizeImageUrl(res.creatorAvatar),
      createTime: res.createTime,
      updateTime: res.updateTime,
      sortOrder: res.sortOrder
    };
  } catch (e) {
    return undefined;
  }
}

export async function addRecipe(name: string, images: string[], content: string): Promise<Recipe> {
  const payloadImages = Array.isArray(images) ? images.filter(Boolean).map(toRelativePath) : [];
  const res = await request<any>('/recipe/add', 'POST', { name, images: payloadImages, content });
  return {
    id: String(res.id),
    familyId: String(res.familyId),
    name: res.name,
    images: res.images,
    content: res.content,
    createTime: res.createTime,
    updateTime: res.updateTime,
    sortOrder: res.sortOrder
  };
}

export async function updateRecipe(id: string, name: string, images: string[], content: string): Promise<void> {
  const payloadImages = Array.isArray(images) ? images.filter(Boolean).map(toRelativePath) : [];
  await request('/recipe/update', 'POST', { id: parseInt(id), name, images: payloadImages, content });
}

export async function deleteRecipe(id: string): Promise<void> {
  await request(`/recipe/delete?id=${id}`, 'POST');
}

export async function reorderRecipe(id: string, direction: 'up' | 'down'): Promise<void> {
  await request('/recipe/reorder', 'POST', { id: parseInt(id), direction });
}

export async function batchUpdateRecipes(updatedRecipes: Recipe[]): Promise<void> {
  const payload = updatedRecipes.map(r => ({
    id: parseInt(r.id),
    sortOrder: r.sortOrder
  }));
  await request('/recipe/batchUpdate', 'POST', { recipes: payload });
}

export interface RecipeTemplate {
  id: string;
  name: string;
  images: string[];
  content: string;
}

export async function getRecipeTemplates(): Promise<RecipeTemplate[]> {
  const res = await request<any[]>('/recipe/templates', 'GET');
  return res.map(t => ({
    id: t.id,
    name: t.name,
    images: Array.isArray(t.images) ? t.images.map(normalizeImageUrl) : [],
    content: t.content
  }));
}
