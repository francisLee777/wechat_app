import { addRecipe, updateRecipe, getRecipeById, deleteRecipe } from '../../utils/db';

Page({
  data: {
    isEdit: false,
    id: '',
    name: '',
    images: [] as string[],
    content: ''
  },

  onLoad(options: any) {
    if (options.id) {
      const recipe = getRecipeById(options.id);
      if (recipe) {
        this.setData({
          isEdit: true,
          id: recipe.id,
          name: recipe.name,
          images: recipe.images || [], // 兼容处理
          content: recipe.content
        });
        wx.setNavigationBarTitle({ title: '编辑食谱' });
      }
    } else {
      wx.setNavigationBarTitle({ title: '新建食谱' });
    }
  },

  // 输入处理
  onNameInput(e: any) { this.setData({ name: e.detail.value }); },
  onContentInput(e: any) { this.setData({ content: e.detail.value }); },

  // 选择图片
  chooseImage() {
    const currentCount = this.data.images.length;
    const maxCount = 3;
    if (currentCount >= maxCount) {
      wx.showToast({ title: '最多上传3张照片', icon: 'none' });
      return;
    }

    wx.chooseMedia({
      count: maxCount - currentCount,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newPaths = res.tempFiles.map(f => f.tempFilePath);
        this.setData({ 
          images: this.data.images.concat(newPaths) 
        });
      }
    });
  },

  // 删除图片
  deleteImage(e: any) {
    const index = e.currentTarget.dataset.index;
    const images = this.data.images;
    images.splice(index, 1);
    this.setData({ images });
  },

  // 预览图片
  previewImage(e: any) {
    const current = e.currentTarget.dataset.src;
    wx.previewImage({
      current,
      urls: this.data.images
    });
  },

  // 保存
  handleSave() {
    const { id, name, images, content, isEdit } = this.data;
    if (!name) {
      wx.showToast({ title: '请输入名称', icon: 'none' });
      return;
    }
    if (!content) {
      wx.showToast({ title: '请输入制作步骤', icon: 'none' });
      return;
    }

    try {
      if (isEdit) {
        updateRecipe(id, name, images, content);
      } else {
        addRecipe(name, images, content);
      }
      wx.showToast({ title: '保存成功', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 1500);
    } catch (e: any) {
      wx.showToast({ title: e.message, icon: 'none' });
    }
  },

  // 删除
  handleDelete() {
    if (!this.data.isEdit) return;
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这个食谱吗？',
      success: (res) => {
        if (res.confirm) {
          try {
            deleteRecipe(this.data.id);
            wx.showToast({ title: '已删除', icon: 'success' });
            setTimeout(() => wx.navigateBack({ delta: 2 }), 1500); // 返回列表
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  }
});
