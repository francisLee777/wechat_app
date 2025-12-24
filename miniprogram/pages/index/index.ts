import { getRecipes, batchUpdateRecipes, getRecipeTemplates, RecipeTemplate } from '../../api/recipe';
import { getCurrentUser } from '../../api/auth';
import type { Recipe } from '../../models/index';

Page({
  data: {
    recipes: [] as Recipe[],
    templates: [] as RecipeTemplate[],
    showTemplateModal: false,
    keyword: '',
    isAdmin: false,
    
    // FAB 拖拽位置
    fabX: 300,
    fabY: 500,

    // 拖拽相关状态
    draggingIndex: -1, // 当前被拖拽的元素索引
    targetIndex: -1,   // 当前拖拽到的目标位置索引
    listStyles: [] as string[], // 存储每个元素的 transform 样式
    
    startY: 0,
    itemHeight: 115, // item高度 + margin，必须与 CSS 保持一致
    swipeXList: [] as number[],
    swipingIndex: -1,
    swipeStartX: 0,
    swipeStartY: 0,
    isSwiping: false,
  },

  async onLoad() {
    // 初始化 FAB 位置 (右下角)
    const { windowWidth, windowHeight } = wx.getSystemInfoSync();
    this.setData({
      fabX: windowWidth - 80, // 预留宽度
      fabY: windowHeight - 150 // 预留高度
    });
  },

  onShow() {
    this.checkRole();
    this.loadData();
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
  },

  checkRole() {
    const user = getCurrentUser();
    this.setData({
      isAdmin: user ? user.role === 'admin' : false
    });
  },

  async loadData() {
    try {
      const list = await getRecipes(this.data.keyword);
      // 初始化样式数组
      const styles = list.map(() => '');
      this.setData({ 
        recipes: list,
        listStyles: styles,
        swipeXList: list.map(() => 0),
        draggingIndex: -1,
        targetIndex: -1
      });
    } catch (e) {
      console.error(e);
    }
  },

  onSearchInput(e: any) {
    this.setData({ keyword: e.detail.value });
    this.loadData();
  },

  goToDetail(e: any) {
    if (this.data.draggingIndex !== -1) return;
    if (this.data.isSwiping) return;
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({ url: `/pages/recipe/detail?id=${id}` });
  },

  goToAdd() {
    wx.navigateTo({ url: '/pages/recipe/edit' });
  },

  async goToAddTemplate() {
    wx.showLoading({ title: '加载模版...' });
    try {
      const templates = await getRecipeTemplates();
      this.setData({
        templates,
        showTemplateModal: true
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    } finally {
      wx.hideLoading();
    }
  },

  closeTemplateModal() {
    this.setData({ showTemplateModal: false });
  },

  selectTemplate(e: any) {
    const index = e.currentTarget.dataset.index;
    const template = this.data.templates[index];
    if (template) {
      // 传递模版数据到编辑页
      // 由于数据可能较长，建议存入本地缓存或事件总线，这里简单起见用 encodeURIComponent
      const tplData = encodeURIComponent(JSON.stringify(template));
      wx.navigateTo({
        url: `/pages/recipe/edit?template=${tplData}`
      });
      this.closeTemplateModal();
    }
  },

  // 拖拽开始
  handleDragStart(e: any) {
    if (!this.data.isAdmin || this.data.keyword) return;
    
    const index = e.currentTarget.dataset.index;
    this.setData({
      draggingIndex: index,
      targetIndex: index,
      startY: e.touches[0].clientY,
      listStyles: this.data.recipes.map((_, i) => i === index ? 'z-index: 100; transition: none;' : 'transition: transform 0.2s;')
    });
  },

  // 拖拽移动 (只改变样式，不改变数据)
  handleDragMove(e: any) {
    if (this.data.draggingIndex === -1) return;

    const currentY = e.touches[0].clientY;
    const diff = currentY - this.data.startY;
    const { draggingIndex, itemHeight, recipes } = this.data;

    // 计算当前的目标索引
    let targetIndex = draggingIndex + Math.round(diff / itemHeight);
    // 限制边界
    targetIndex = Math.max(0, Math.min(targetIndex, recipes.length - 1));

    const styles: string[] = [];
    
    for (let i = 0; i < recipes.length; i++) {
      if (i === draggingIndex) {
        // 被拖拽的元素：跟随手指移动 (加上缩放效果)
        styles.push(`transform: translateY(${diff}px) scale(1.05); z-index: 100; transition: none; box-shadow: 0 10px 20px rgba(0,0,0,0.2);`);
      } else {
        // 其他元素：计算是否需要让位
        let offset = 0;
        
        // 如果被拖拽元素往下移 (dragging < target)
        if (draggingIndex < targetIndex) {
          // 在区间内的元素往上移
          if (i > draggingIndex && i <= targetIndex) {
            offset = -itemHeight;
          }
        } 
        // 如果被拖拽元素往上移 (dragging > target)
        else if (draggingIndex > targetIndex) {
          // 在区间内的元素往下移
          if (i >= targetIndex && i < draggingIndex) {
            offset = itemHeight;
          }
        }
        
        styles.push(offset !== 0 ? `transform: translateY(${offset}px); transition: transform 0.2s;` : 'transition: transform 0.2s;');
      }
    }

    this.setData({
      listStyles: styles,
      targetIndex: targetIndex
    });
  },

  // 拖拽结束 (延迟交换，实现平滑归位)
  handleDragEnd() {
    const { draggingIndex, targetIndex, recipes, itemHeight } = this.data;
    if (draggingIndex === -1) return;

    // 1. 计算被拖拽元素的最终物理位置偏移量
    const finalOffset = (targetIndex - draggingIndex) * itemHeight;

    // 2. 更新样式：让被拖拽元素平滑飞到目标位置 (应用 transition)
    // 注意：其他已让位的元素样式保持不变，等待数据交换后重置
    const newStyles = [...this.data.listStyles];
    newStyles[draggingIndex] = `transform: translateY(${finalOffset}px); z-index: 100; transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1); box-shadow: 0 2px 8px rgba(0,0,0,0.1);`;

    this.setData({ listStyles: newStyles });

    // 3. 等待动画结束后，交换数据并重置
    setTimeout(() => {
      // 必须将数据更新和样式重置合并在同一个 setData 中，防止闪烁
      let newRecipes = [...recipes]; // 总是基于最新数据操作
      
      if (draggingIndex !== targetIndex) {
        const temp = newRecipes[draggingIndex];
        newRecipes.splice(draggingIndex, 1);
        newRecipes.splice(targetIndex, 0, temp);

        // 更新数据库 sortOrder
        const len = newRecipes.length;
        const now = Date.now();
        const updatedRecipes = newRecipes.map((r, i) => ({
          ...r,
          sortOrder: now + (len - i) * 1000
        }));

        batchUpdateRecipes(updatedRecipes).catch(err => console.error(err));
      }

      // 原子操作：同时更新数据和重置样式
      this.setData({
        recipes: newRecipes,
        draggingIndex: -1,
        targetIndex: -1,
        listStyles: newRecipes.map(() => '') // 样式全部清零
      });
    }, 300); // 稍微延长等待时间，确保动画完全结束
  }
  ,
  handleItemTouchStart(e: any) {
    const index = e.currentTarget.dataset.index;
    const x = e.touches[0].clientX;
    const y = e.touches[0].clientY;
    this.setData({ swipingIndex: index, swipeStartX: x, swipeStartY: y, isSwiping: false });
  },
  handleItemTouchMove(e: any) {
    const { swipingIndex, swipeStartX, swipeStartY, swipeXList } = this.data as any;
    if (swipingIndex === -1) return;
    const x = e.touches[0].clientX;
    const y = e.touches[0].clientY;
    const dx = x - swipeStartX;
    const dy = y - swipeStartY;
    if (!this.data.isSwiping) {
      if (Math.abs(dx) > Math.abs(dy) && Math.abs(dx) > 5) {
        this.setData({ isSwiping: true });
      } else {
        return;
      }
    }
    let nx = swipeXList[swipingIndex] + dx;
    if (nx > 0) nx = 0;
    const minX = -80;
    if (nx < minX) nx = minX;
    swipeXList[swipingIndex] = nx;
    this.setData({ swipeXList, swipeStartX: x, swipeStartY: y });
  },
  handleItemTouchEnd() {
    const { swipingIndex, swipeXList } = this.data as any;
    if (swipingIndex === -1) return;
    const threshold = -40;
    const snapOpen = -80;
    const snapClose = 0;
    const nx = swipeXList[swipingIndex] <= threshold ? snapOpen : snapClose;
    swipeXList[swipingIndex] = nx;
    this.setData({ swipeXList, swipingIndex: -1, isSwiping: false });
  },
  async deleteRecipeQuick(e: any) {
    const id = e.currentTarget.dataset.id;
    const name = e.currentTarget.dataset.name;
    const that = this;
    wx.showModal({
      title: '删除确认',
      content: `确定要删除“${name}”吗？`,
      confirmText: '删除',
      confirmColor: '#fa5151',
      success: async (res) => {
        if (res.confirm) {
          try {
            const { deleteRecipe } = require('../../api/recipe');
            await deleteRecipe(id);
            const list = that.data.recipes.filter((r: Recipe) => r.id !== id);
            that.setData({ recipes: list, swipeXList: list.map(() => 0) });
            wx.showToast({ title: '已删除', icon: 'success' });
          } catch (err: any) {
            wx.showToast({ title: err.message || '删除失败', icon: 'none' });
          }
        }
      }
    });
  }
});
