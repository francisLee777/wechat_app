import { joinFamily } from '../../utils/db';

Page({
  data: {
    familyId: ''
  },

  onIdInput(e: any) {
    this.setData({
      familyId: e.detail.value
    });
  },

  async handleJoin() {
    const id = this.data.familyId.trim();
    if (!id) {
      wx.showToast({ title: '请输入家庭ID', icon: 'none' });
      return;
    }

    try {
      const family = await joinFamily(id);
      wx.showToast({ title: `成功加入: ${family.name}`, icon: 'success' });
      
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
    } catch (error: any) {
      wx.showToast({ title: error.message, icon: 'none' });
    }
  }
});
