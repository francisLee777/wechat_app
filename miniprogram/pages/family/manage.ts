import { getFamilyMembers, removeMember, quitFamily, getFamilyById, deleteFamily } from '../../api/family';
import { getCurrentUser } from '../../api/auth';
import type { User, Family } from '../../models/index';

Page({
  data: {
    family: null as Family | null,
    members: [] as User[],
    currentUser: null as any
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({ selected: 3 });
    }
    wx.showTabBar({});
    this.loadData();
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
  },

  async loadData() {
    const user = getCurrentUser();
    this.setData({ currentUser: user });

    if (user && user.familyId) {
      try {
        const family = await getFamilyById(user.familyId);
        const members = await getFamilyMembers(user.familyId);
        this.setData({
          family: family || null,
          members: members
        });
      } catch (e) {
        console.error(e);
      }
    } else {
      // 未加入家庭，仅清空数据，等待 UI 渲染空状态
      this.setData({
        family: null,
        members: []
      });
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
      success: async (res) => {
        if (res.confirm) {
          try {
            await removeMember(userId);
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
      success: async (res) => {
        if (res.confirm) {
          try {
            await quitFamily();
            // 退出后刷新当前页，显示“空状态”
            this.loadData();
            wx.showToast({ title: '已退出', icon: 'success' });
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  },

  // 解散家庭
  handleDelete() {
    wx.showModal({
      title: '危险操作',
      content: '确定要解散当前家庭吗？此操作将删除所有留言、食谱且不可恢复！',
      confirmColor: '#FF0000',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteFamily();
            wx.showToast({ title: '家庭已解散', icon: 'none' });
            // 解散后刷新当前页，显示“空状态”
            this.loadData();
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  }
});
