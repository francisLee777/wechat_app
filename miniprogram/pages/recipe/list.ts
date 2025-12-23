import { getRecipes, Recipe, getCurrentUser, batchUpdateRecipes } from '../../utils/db';

Page({
  data: {
    recipes: [] as Recipe[],
    keyword: '',
    isAdmin: false,
    
    // 拖拽相关状态
    draggingIndex: -1, // 当前被拖拽的元素索引
    targetIndex: -1,   // 当前拖拽到的目标位置索引
    listStyles: [] as string[], // 存储每个元素的 transform 样式
    
    startY: 0,
    itemHeight: 115, // item高度 + margin，必须与 CSS 保持一致
  },

  onShow() {
    this.checkRole();
    this.loadData();
  },

  checkRole() {
    const user = getCurrentUser();
    this.setData({
      isAdmin: user.role === 'admin'
    });
  },

  loadData() {
    const list = getRecipes(this.data.keyword);
    // 初始化样式数组
    const styles = list.map(() => '');
    this.setData({ 
      recipes: list,
      listStyles: styles,
      draggingIndex: -1,
      targetIndex: -1
    });
  },

  onSearchInput(e: any) {
    this.setData({ keyword: e.detail.value });
    this.loadData();
  },

  goToDetail(e: any) {
    // 拖拽时不跳转
    if (this.data.draggingIndex !== -1) return;
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({ url: `/pages/recipe/detail?id=${id}` });
  },

  goToAdd() {
    wx.navigateTo({ url: '/pages/recipe/edit' });
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

        batchUpdateRecipes(updatedRecipes);
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
});
