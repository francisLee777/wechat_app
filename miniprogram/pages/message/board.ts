import { getMessages, addMessage, deleteMessage } from '../../api/message';
import { getCurrentUser } from '../../api/auth';
import type { Message } from '../../models/index';
import { formatTime } from '../../utils/util';

Page({
  data: {
    messages: [] as any[],
    content: '',
    currentUser: null as any
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({ selected: 2 });
    }
    wx.showTabBar({});
    this.setData({ currentUser: getCurrentUser() });
    this.loadData();
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
  },

  async loadData() {
    try {
      const rawList = await getMessages();
      const list = rawList.map(msg => ({
        ...msg,
        timeStr: formatTime(new Date(msg.createTime)),
        replies: (msg.replies || []).map((r: any) => ({
          ...r,
          timeStr: formatTime(new Date(r.createTime))
        }))
      }));
      this.setData({ messages: list });
    } catch (e) {
      console.error(e);
    }
  },

  onInput(e: any) {
    this.setData({ content: e.detail.value });
  },

  // 发送留言
  async handleSend() {
    if (!this.data.content.trim()) return;
    try {
      await addMessage(this.data.content);
      this.setData({ content: '' });  
      this.loadData();
      wx.showToast({ title: '已发送', icon: 'none' });
    } catch (e: any) {
      wx.showToast({ title: e.message, icon: 'none' });
    } 
  },

  // 删除留言 (根留言或回复)
  handleDelete(e: any) { 
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定删除吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteMessage(id);
            this.loadData();
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  },

  // 回复留言
  handleReply(e: any) {
    const parentId = e.currentTarget.dataset.id;
    const replyToUser = e.currentTarget.dataset.user;
    
    wx.showModal({
      title: `回复 ${replyToUser}`,
      editable: true,
      placeholderText: '请输入回复内容',
      success: async (res) => {
        if (res.confirm && res.content) {
          if (res.content.length > 100) {
            wx.showToast({ title: '回复内容不能超过100字', icon: 'none' });
            return;
          }
          try {
            // 调用统一的 addMessage 接口，传入 parentId
            await addMessage(res.content, parentId);
            this.loadData();
          } catch (e: any) {
            wx.showToast({ title: e.message, icon: 'none' });
          }
        }
      }
    });
  }
});
