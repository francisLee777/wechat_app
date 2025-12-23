import { createFamily } from '../../utils/db';

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
  handleCreate() {
    const name = this.data.familyName.trim();
    if (!name) {
      wx.showToast({ title: '请输入家庭名称', icon: 'none' });
      return;
    }

    try {
      const family = createFamily(name);
      wx.showToast({ title: '创建成功', icon: 'success' });
      
      // 延迟返回首页
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
    } catch (error: any) {
      wx.showToast({ title: error.message, icon: 'none' });
    }
  }
});
