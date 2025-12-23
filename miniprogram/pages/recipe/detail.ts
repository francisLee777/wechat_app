import { getRecipeById, getCurrentUser, Recipe } from '../../utils/db';

Page({
  data: {
    recipe: null as Recipe | null,
    isAdmin: false,
    currentImageIndex: 0
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
  },

  async loadData(id: string) {
    const recipe = await getRecipeById(id);
    const user = getCurrentUser();
    
    this.setData({
      recipe: recipe || null,
      isAdmin: user ? user.role === 'admin' : false,
      currentImageIndex: 0 // 重置图片索引
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
      wx.previewImage({
        current,
        urls: this.data.recipe.images
      });
    }
  }
});
