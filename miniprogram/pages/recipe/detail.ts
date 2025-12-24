import { getRecipeById, deleteRecipe } from '../../api/recipe';
import { getCurrentUser } from '../../api/auth';
import type { Recipe } from '../../models/index';

Page({
  data: {
    recipe: null as Recipe | null,
    isAdmin: false,
    currentImageIndex: 0,
    updateTimeStr: ''
  },

  onLoad(options: any) {
    if (options.id) {
      this.loadData(options.id);
    }
  },

  onShow() {
    // 每次显示都刷新，因为可能从编辑页回来
    if (this.data.recipe) {
      this.loadData(this.data.recipe.id);
    }
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
  },

  async loadData(id: string) {
    const recipe = await getRecipeById(id);
    const user = getCurrentUser();
    let updateTimeStr = '';
    if (recipe && recipe.updateTime) {
      const d = new Date(recipe.updateTime);
      const y = d.getFullYear();
      const m = (d.getMonth() + 1).toString().padStart(2, '0');
      const day = d.getDate().toString().padStart(2, '0');
      const hh = d.getHours().toString().padStart(2, '0');
      const mm = d.getMinutes().toString().padStart(2, '0');
      updateTimeStr = `${y}-${m}-${day} ${hh}:${mm}`;
    }
    
    this.setData({
      recipe: recipe || null,
      isAdmin: user ? user.role === 'admin' : false,
      currentImageIndex: 0,
      updateTimeStr
    });
  },

  goToEdit() {
    if (this.data.recipe) {
      wx.navigateTo({ url: `/pages/recipe/edit?id=${this.data.recipe.id}` });
    }
  },

  // 监听轮播图切换
  onSwiperChange(e: any) {
    this.setData({
      currentImageIndex: e.detail.current
    });
  },

  // 预览图片
  previewImage(e: any) {
    const current = e.currentTarget.dataset.current;
    if (this.data.recipe && this.data.recipe.images) {
      const token = wx.getStorageSync('AUTH_TOKEN');
      const base = require('../../utils/request');
      const { IMG_BASE_URL } = base;
      const urls = this.data.recipe.images.map((p: string) => {
        const full = `${IMG_BASE_URL}${p.startsWith('/') ? p : '/' + p}`;
        const sep = full.indexOf('?') >= 0 ? '&' : '?';
        return token ? `${full}${sep}token=${token}` : full;
      });
      wx.previewImage({
        current,
        urls
      });
    }
  },

  // 快捷删除
  async handleDelete() {
    if (!this.data.isAdmin || !this.data.recipe) return;
    const id = this.data.recipe.id;
    wx.showModal({
      title: '删除确认',
      content: '确定要删除此食谱吗？该操作不可恢复',
      confirmText: '删除',
      confirmColor: '#fa5151',
      cancelText: '取消',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteRecipe(id);
            wx.showToast({ title: '已删除', icon: 'success' });
            wx.navigateBack();
          } catch (e: any) {
            wx.showToast({ title: e.message || '删除失败', icon: 'none' });
          }
        }
      }
    });
  }
});
