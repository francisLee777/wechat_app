import { createFamily } from '../../api/family';

Page({
  data: {
    familyName: ''
  },

  // 输入框变化处理
  onNameInput(e: any) {
    this.setData({
      familyName: e.detail.value
    });
  },

  // 创建家庭
  async handleCreate() {
    const name = this.data.familyName.trim();
    if (!name) {
      wx.showToast({ title: '请输入家庭名称', icon: 'none' });
      return;
    }
    if (name.length > 20) {
      wx.showToast({ title: '家庭名称不能超过20字', icon: 'none' });
      return;
    }

    try {
      const family = await createFamily(name);
      wx.showToast({ title: '创建成功', icon: 'success' });
      
      // 延迟返回首页
      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' });
      });
    } catch (error: any) {
      wx.showToast({ title: error.message, icon: 'none' });
    }
  }
});
