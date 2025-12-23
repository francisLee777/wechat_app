import { getCurrentUser, User } from '../../utils/db';

Page({
  data: {
    hasFamily: false,
    user: null as User | null,
  },

  onShow() {
    this.refreshUser();
  },

  refreshUser() {
    const user = getCurrentUser();
    this.setData({
      user,
      hasFamily: !!user.familyId
    });
  },

  // 跳转到创建家庭页面
  goToCreateFamily() {
    wx.navigateTo({ url: '/pages/family/create' });
  },

  // 跳转到加入家庭页面
  goToJoinFamily() {
    wx.navigateTo({ url: '/pages/family/join' });
  },

  // 跳转到食谱列表
  goToRecipes() {
    wx.navigateTo({ url: '/pages/recipe/list' });
  },

  // 跳转到留言板
  goToMessages() {
    wx.navigateTo({ url: '/pages/message/board' });
  },

  // 跳转到家庭管理
  goToManage() {
    wx.navigateTo({ url: '/pages/family/manage' });
  },
  
  // 复制用户ID (方便测试)
  copyUserId() {
    if (this.data.user) {
      wx.setClipboardData({
        data: this.data.user.id,
        success: () => wx.showToast({ title: '用户ID已复制' })
      });
    }
  }
});
