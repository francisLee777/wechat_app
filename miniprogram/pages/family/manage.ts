import { getFamilyMembers, removeMember, quitFamily, getCurrentUser, getFamilyById, User, Family } from '../../utils/db';

Page({
  data: {
    family: null as Family | null,
    members: [] as User[],
    currentUser: null as any
  },

  onShow() {
    this.loadData();
  },

  loadData() {
    const user = getCurrentUser();
    this.setData({ currentUser: user });

    if (user.familyId) {
      const family = getFamilyById(user.familyId);
      const members = getFamilyMembers(user.familyId);
      this.setData({
        family: family || null,
        members: members
      });
    } else {
      // 异常情况，返回首页
      wx.navigateBack();
    }
  },

  // 复制家庭ID
  copyFamilyId() {
    if (this.data.family) {
      wx.setClipboardData({
        data: this.data.family.id,
        success: () => wx.showToast({ title: '邀请码已复制' })
      });
    }
  },

  // 移除成员
  handleRemove(e: any) {
    const userId = e.currentTarget.dataset.id;
    const name = e.currentTarget.dataset.name;

    wx.showModal({
      title: '确认移除',
      content: `确定要将 ${name} 移出家庭吗？`,
      success: (res) => {
        if (res.confirm) {
          try {
            removeMember(userId);
            this.loadData(); // 刷新列表
            wx.showToast({ title: '已移除', icon: 'success' });
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  },

  // 退出家庭
  handleQuit() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出当前家庭吗？',
      success: (res) => {
        if (res.confirm) {
          try {
            quitFamily();
            wx.reLaunch({ url: '/pages/index/index' }); // 退出后重置到首页
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  }
});
