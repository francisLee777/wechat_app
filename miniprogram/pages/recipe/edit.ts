import { addRecipe, updateRecipe, getRecipeById, deleteRecipe } from '../../api/recipe';
import { uploadFile, IMG_BASE_URL } from '../../utils/request';

Page({
  data: {
    isEdit: false,
    id: '',
    name: '',
    images: [] as string[],
    content: ''
  },

  async onLoad(options: any) {
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
    if (options.id) {
      const recipe = await getRecipeById(options.id);
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
    } else if (options.template) {
      // 从模版导入
      try {
        const tpl = JSON.parse(decodeURIComponent(options.template));
        
        // 处理模版图片：下载到本地临时文件，以便后续 handleSave 时能触发上传逻辑
        let localImages: string[] = [];
        if (tpl.images && tpl.images.length > 0) {
           wx.showLoading({ title: '准备模版资源...' });
           try {
             localImages = await Promise.all(tpl.images.map((imgUrl: string) => {
               // 拼接完整 URL (如果是相对路径)
                let fullUrl = imgUrl;
                if (!fullUrl.startsWith('http')) {
                   fullUrl = IMG_BASE_URL + imgUrl;
                }
               
               return new Promise<string>((resolve, reject) => {
                 wx.downloadFile({
                   url: fullUrl,
                   header: { 'Authorization': `Bearer ${token}` }, // 模版图片可能需要鉴权
                   success: (res) => {
                     if (res.statusCode === 200) {
                       resolve(res.tempFilePath);
                     } else {
                       console.warn('Download template image failed', res);
                       resolve(''); // 失败则忽略该图
                     }
                   },
                   fail: (err) => {
                     console.warn('Download template image error', err);
                     resolve('');
                   }
                 });
               });
             }));
             // 过滤掉下载失败的空字符串
             localImages = localImages.filter(p => p);
           } catch (e) {
             console.error("Download template images error", e);
           } finally {
             wx.hideLoading();
           }
        }

        this.setData({
          isEdit: false,
          name: tpl.name,
          images: localImages,
          content: tpl.content
        });
        wx.setNavigationBarTitle({ title: '新建食谱 (模版)' });
      } catch (e) {
        console.error("Parse template failed", e);
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
        console.log("选择图片后 ",newPaths)
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
  async handleSave() {
    let { id, name, images, content, isEdit } = this.data;
    if (!name) {
      wx.showToast({ title: '请输入名称', icon: 'none' });
      return;
    }
    if (name.length > 50) {
      wx.showToast({ title: '名称不能超过50字', icon: 'none' });
      return;
    }
    if (!content) {
      wx.showToast({ title: '请输入制作步骤', icon: 'none' });
      return;
    }
    if (content.length > 2000) {
      wx.showToast({ title: '内容不能超过2000字', icon: 'none' });
      return;
    }

    try {
      wx.showLoading({ title: '保存中...' });

      // 上传新添加的图片 (临时路径)
      const uploadedImages = await Promise.all(images.map(async (img) => {
        // 如果是本地临时文件（http://tmp 或 wxfile://），则上传
        if (img.startsWith('http://tmp') || img.startsWith('wxfile://')) {
          try {
            // 默认 scope 为 family，因为食谱是家庭共享的
            const serverUrl = await uploadFile(img, 'family');
            return serverUrl;
          } catch (e) {
            console.error('Upload failed for', img, e);
            throw new Error('图片上传失败，请重试');
          }
        }
        // 已经是服务器 URL，直接返回
        return img;
      }));

      // 更新图片列表
      images = uploadedImages;

      if (isEdit) {
        await updateRecipe(id, name, images, content);
      } else {
        await addRecipe(name, images, content);
      }
      wx.hideLoading();
      wx.showToast({ title: '保存成功', icon: 'success' });
      setTimeout(() => wx.navigateBack(), 1500);
    } catch (e: any) {
      wx.hideLoading();
      wx.showToast({ title: e.message, icon: 'none' });
    }
  },

  // 删除
  handleDelete() {
    if (!this.data.isEdit) return;
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这个食谱吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteRecipe(this.data.id);
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
