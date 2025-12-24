import { getCurrentUser, loginUser, checkSession, saveNickName, saveAvatarUrl } from '../../api/auth';
import { uploadFile } from '../../utils/request';
import { joinFamily } from '../../api/family';
import type { User } from '../../models/index';

Page({
  data: {
    isLoggedIn: false,
    hasFamily: false,
    user: null as User | null,
    tempNickName: ''
    ,isProfileRequired: false
  },

  async onShow() {
    // 检查会话是否有效
    const isValid = await checkSession();
    if (isValid) {
      this.refreshUser();
    } else {
      this.setData({
        isLoggedIn: false,
        user: null,
        hasFamily: false
      });
    }
    // Set token for WXS image loading
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
  },

  refreshUser() {
    const user = getCurrentUser();
    const needsNick = !user || !user.nickName || user.nickName.trim() === '' || user.nickName === '微信用户';
    // const needsAvatar = !user || !user.avatarUrl || user.avatarUrl.trim() === '';
    this.setData({
      isLoggedIn: !!user,
      user,
      hasFamily: !!(user && user.familyId),
      tempNickName: user ? user.nickName : '',
      isProfileRequired: needsNick,
    });
  },

  // 登录
    async handleLogin() {
    const that = this;
    wx.showLoading({ title: '登录中...' });
    loginUser().then(() => {
      // 登录成功后，立即更新页面数据中的 token
      const token = wx.getStorageSync('AUTH_TOKEN');
      that.setData({ token });
      
      that.refreshUser();
      if (that.data.isProfileRequired) {
        wx.showToast({ title: '请先完善头像与昵称', icon: 'none' });
      } else {
        wx.showToast({ title: '登录成功', icon: 'success' });
        // 登录成功且已加入家庭，自动跳转到业务首页
        if (that.data.hasFamily) {
          console.log("登录成功且已加入家庭，自动跳转到业务首页")
          // 立即跳转，无需等待
          wx.reLaunch({ url: '/pages/index/index' });
        }
      }
      wx.hideLoading();
    }).catch(() => {
      wx.hideLoading();
      wx.showToast({ title: '登录失败', icon: 'none' });
    });
  },

  // 处理头像选择
  async onChooseAvatar(e: any) {
    const { avatarUrl } = e.detail;
    if (!avatarUrl) return;

    try {
      wx.showLoading({ title: '上传中...' });
      // 1. 上传文件到服务器
      const remoteUrl = await uploadFile(avatarUrl, 'user');
      
      // 2. 保存 URL 到用户信息
      await saveAvatarUrl(remoteUrl);
      
      // 3. 刷新本地用户状态
      this.refreshUser();
      
      wx.hideLoading();
      wx.showToast({ title: '头像更新成功', icon: 'success' });
    } catch (err: any) {
      wx.hideLoading();
      wx.showToast({ title: '头像更新失败', icon: 'none' });
      console.error(err);
    }
  },

  // 昵称输入（仅更新本地临时值，不触发后端保存）
  onNickNameInput(e: any) {
    this.setData({ tempNickName: e.detail.value });
  },

  // 昵称保存（失焦或确认时触发）
  async onNickNameSave(e: any) {
    // 优先从事件中获取最新值，因为 setData 可能有延迟
    let nickName = e.detail.value;
    // 如果事件中没有值（极端情况），回退到 data
    if (nickName === undefined || nickName === null) {
      nickName = this.data.tempNickName;
    }
    
    // 去除首尾空格
    nickName = String(nickName).trim();
    
    // 更新 tempNickName 以保持 UI 同步
    if (nickName !== this.data.tempNickName) {
       this.setData({ tempNickName: nickName });
    }

    const current = this.data.user ? this.data.user.nickName : '';
    // 如果为空，或者与当前已保存的昵称相同，或者仍为默认值“微信用户”且用户未修改，则不保存
    if (!nickName || nickName === current) return;

    try {
      wx.showLoading({ title: '保存中...' });
      await saveNickName(nickName);
      // 重新获取用户信息以更新全局状态
      const updatedUser = await getCurrentUser(); // 这里应该 fetchCurrentUser 但 refreshUser 内部只是 getCurrentUser
      // 应该调用 fetchCurrentUser 强制刷新缓存
      // 这里为了简单，直接复用 refreshUser 的逻辑，但 refreshUser 依赖缓存。
      // 所以我们手动调用一次 fetchCurrentUser 更新缓存
      const { fetchCurrentUser } = require('../../api/auth');
      await fetchCurrentUser();
      
      this.refreshUser();
      wx.hideLoading();
      wx.showToast({ title: '昵称更新成功', icon: 'success' });

      // 昵称保存成功且已加入家庭，自动跳转到业务首页
      if (this.data.hasFamily) {
        // 立即跳转
        wx.reLaunch({ url: '/pages/index/index' });
      }
    } catch (err: any) {
      wx.hideLoading();
      wx.showToast({ title: '昵称更新失败', icon: 'none' });
      console.error(err);
      // 恢复为原值
      this.setData({ tempNickName: current });
    }
  },

  // 跳转到创建家庭页面
  goToCreateFamily() {
    if (this.data.isProfileRequired) { wx.showToast({ title: '请先完善头像与昵称', icon: 'none' }); return; }
    wx.navigateTo({ url: '/pages/family/create' });
  },

  // 跳转到加入家庭页面
  goToJoinFamily() {
    if (this.data.isProfileRequired) { wx.showToast({ title: '请先完善头像与昵称', icon: 'none' }); return; }
    wx.navigateTo({ url: '/pages/family/join' });
  },

  // 跳转到食谱列表
  goToRecipes() {
    if (this.data.isProfileRequired) { wx.showToast({ title: '请先完善头像与昵称', icon: 'none' }); return; }
    wx.switchTab({ url: '/pages/index/index' });
  },

  // 跳转到留言板
  goToMessages() {
    if (this.data.isProfileRequired) { wx.showToast({ title: '请先完善头像与昵称', icon: 'none' }); return; }
    wx.switchTab({ url: '/pages/message/board' });
  },

  // 跳转到家庭管理
  goToManage() {
    if (this.data.isProfileRequired) { wx.showToast({ title: '请先完善头像与昵称', icon: 'none' }); return; }
    wx.switchTab({ url: '/pages/family/manage' });
  },
});
