import { getMenuList, addMenu, deleteMenu, MenuListResponse } from '../../api/menu';
import { getRecipes } from '../../api/recipe';
import { Recipe } from '../../models/index';
import { formatTime } from '../../utils/util';

Page({
  data: {
    baseDate: '',
    days: [] as Array<{
      date: string,
      label: string,
      weekDay: string,
      menu: MenuListResponse,
      swipeXBreakfast: number[],
      swipeXLunch: number[],
      swipeXDinner: number[],
    }>,
    
    // 弹窗相关
    showAddModal: false,
    currentMealType: 1, // 1,2,3
    currentMealName: '',
    currentMealDate: '',
    
    recipes: [] as Recipe[], // 原始列表
    filteredRecipes: [] as Recipe[], // 搜索后列表
    keyword: '',
    
    selectedRecipeIds: [] as string[],
    remark: '',
    
    token: '',

    // 左滑删除状态（针对某个天的某个餐别）
    swipingDayIndex: -1,
    swipingKey: '', // breakfast|lunch|dinner
    swipingIndex: -1,
    swipeStartX: 0,
    swipeStartY: 0,
    isSwiping: false
  },

  onLoad() {
    const now = new Date();
    const y = now.getFullYear();
    const m = (now.getMonth() + 1).toString().padStart(2, '0');
    const d = now.getDate().toString().padStart(2, '0');
    const base = `${y}-${m}-${d}`;
    this.setBaseDate(base);
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({ selected: 1 });
    }
    wx.showTabBar({});
    const token = wx.getStorageSync('AUTH_TOKEN');
    this.setData({ token });
    this.loadMenus();
  },

  // 初始化三天（起始/次日/后日）
  initDays(baseDateStr: string) {
    const weeks = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    const labels = ['今天', '明天', '后天'];
    const days: any[] = [];
    const base = new Date(baseDateStr);
    for (let i = 0; i < 3; i++) {
      const d = new Date(base);
      d.setDate(base.getDate() + i);
      const y = d.getFullYear();
      const m = (d.getMonth() + 1).toString().padStart(2, '0');
      const day = d.getDate().toString().padStart(2, '0');
      const dateStr = `${y}-${m}-${day}`;
      const weekDay = weeks[d.getDay()];
      days.push({
        date: dateStr,
        label: labels[i],
        weekDay,
        menu: { breakfast: [], lunch: [], dinner: [] },
        swipeXBreakfast: [],
        swipeXLunch: [],
        swipeXDinner: [],
      });
    }
    this.setData({ baseDate: baseDateStr, days });
  },

  setBaseDate(baseDateStr: string) {
    this.initDays(baseDateStr);
    this.loadMenus();
  },

  onBaseDateChange(e: any) {
    const dateStr = e.detail.value;
    this.setBaseDate(dateStr);
  },

  prevChunk() {
    const cur = new Date(this.data.baseDate);
    cur.setDate(cur.getDate() - 3);
    const y = cur.getFullYear();
    const m = (cur.getMonth() + 1).toString().padStart(2, '0');
    const d = cur.getDate().toString().padStart(2, '0');
    this.setBaseDate(`${y}-${m}-${d}`);
  },

  nextChunk() {
    const cur = new Date(this.data.baseDate);
    cur.setDate(cur.getDate() + 3);
    const y = cur.getFullYear();
    const m = (cur.getMonth() + 1).toString().padStart(2, '0');
    const d = cur.getDate().toString().padStart(2, '0');
    this.setBaseDate(`${y}-${m}-${d}`);
  },

  // 加载三天菜单
  async loadMenus() {
    try {
      wx.showLoading({ title: '加载中...' });
      const promises = this.data.days.map(d => getMenuList(d.date));
      const results = await Promise.all(promises);
      const days = this.data.days.map((d, i) => ({
        ...d,
        menu: results[i],
        swipeXBreakfast: results[i].breakfast.map(() => 0),
        swipeXLunch: results[i].lunch.map(() => 0),
        swipeXDinner: results[i].dinner.map(() => 0),
      }));
      this.setData({ days, swipingDayIndex: -1, swipingKey: '', swipingIndex: -1, isSwiping: false });
    } catch (e) {
      console.error(e);
    } finally {
      wx.hideLoading();
    }
  },

  

  // 打开添加弹窗
  async openAddModal(e: any) {
    const type = parseInt(e.currentTarget.dataset.type);
    const date = String(e.currentTarget.dataset.date);
    const names = ['', '早餐', '午餐', '晚餐'];
    
    this.setData({
      showAddModal: true,
      currentMealType: type,
      currentMealName: names[type],
      currentMealDate: date,
      keyword: '',
      selectedRecipeIds: [],
      remark: ''
    });

    // 加载食谱库（如果还没加载过）
    if (this.data.recipes.length === 0) {
      try {
        const list = await getRecipes('');
        this.setData({ 
          recipes: list,
          filteredRecipes: list // 初始显示全部
        });
      } catch (e) {
        wx.showToast({ title: '食谱加载失败', icon: 'none' });
      }
    } else {
        // 重置搜索列表
        this.setData({ filteredRecipes: this.data.recipes });
    }
  },

  closeAddModal() {
    this.setData({ showAddModal: false });
  },

  // 搜索食谱
  onSearchInput(e: any) {
    const keyword = e.detail.value;
    const list = this.data.recipes.filter(r => r.name.includes(keyword));
    this.setData({
      keyword,
      filteredRecipes: list
    });
  },

  // 选择食谱（支持多选）
  selectRecipe(e: any) {
    const id = String(e.currentTarget.dataset.id);
    const { selectedRecipeIds } = this.data;
    const index = selectedRecipeIds.indexOf(id);
    
    if (index > -1) {
      // 已选中，移除
      selectedRecipeIds.splice(index, 1);
    } else {
      // 未选中，添加
      selectedRecipeIds.push(id);
    }
    
    this.setData({ selectedRecipeIds });
  },

  // 输入备注
  onRemarkInput(e: any) {
    this.setData({ remark: e.detail.value });
  },

  // 确认添加
  async confirmAdd() {
    if (this.data.selectedRecipeIds.length === 0) return;
    
    try {
      wx.showLoading({ title: '添加中...' });
      
      const tasks = this.data.selectedRecipeIds.map(id => 
        addMenu(
          this.data.currentMealDate,
          this.data.currentMealType,
          id,
          this.data.remark
        )
      );

      await Promise.all(tasks);
      
      wx.showToast({ title: `已添加${tasks.length}道菜`, icon: 'success' });
      this.closeAddModal();
      this.loadMenus(); // 刷新列表
    } catch (e: any) {
      wx.showToast({ title: e.message || '部分添加失败', icon: 'none' });
      // 即使部分失败，也刷新一下看看
      this.loadMenus();
    } finally {
      wx.hideLoading();
    }
  },

  // 删除菜品
  deleteDish(e: any) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '提示',
      content: '确定移除这道菜吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteMenu(id);
            this.loadMenus();
          } catch (e: any) {
            wx.showToast({ title: '删除失败', icon: 'none' });
          }
        }
      }
    });
  }
  ,
  // 左滑删除交互
  handleDishTouchStart(e: any) {
    const dayIndex = parseInt(e.currentTarget.dataset.dayindex);
    const key = e.currentTarget.dataset.section; // breakfast/lunch/dinner
    const index = e.currentTarget.dataset.index;
    const x = e.touches[0].clientX;
    const y = e.touches[0].clientY;
    this.setData({ swipingDayIndex: dayIndex, swipingKey: key, swipingIndex: index, swipeStartX: x, swipeStartY: y, isSwiping: false });
  },
  handleDishTouchMove(e: any) {
    const { swipingDayIndex, swipingKey, swipingIndex, swipeStartX, swipeStartY } = this.data as any;
    if (swipingDayIndex === -1 || !swipingKey || swipingIndex === -1) return;
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
    const minX = -80; // 最大左滑距离
    const day = this.data.days[swipingDayIndex];
    let arr = swipingKey === 'breakfast' ? day.swipeXBreakfast : (swipingKey === 'lunch' ? day.swipeXLunch : day.swipeXDinner);
    let nx = arr[swipingIndex] + dx;
    if (nx > 0) nx = 0;
    if (nx < minX) nx = minX;
    arr[swipingIndex] = nx;
    const days = [...this.data.days];
    if (swipingKey === 'breakfast') days[swipingDayIndex].swipeXBreakfast = arr;
    else if (swipingKey === 'lunch') days[swipingDayIndex].swipeXLunch = arr;
    else days[swipingDayIndex].swipeXDinner = arr;
    this.setData({ days, swipeStartX: x, swipeStartY: y });
  },
  handleDishTouchEnd() {
    const { swipingDayIndex, swipingKey, swipingIndex } = this.data as any;
    if (swipingDayIndex === -1 || !swipingKey || swipingIndex === -1) return;
    const day = this.data.days[swipingDayIndex];
    const arr = swipingKey === 'breakfast' ? day.swipeXBreakfast : (swipingKey === 'lunch' ? day.swipeXLunch : day.swipeXDinner);
    const threshold = -40;
    const snapOpen = -80;
    const snapClose = 0;
    arr[swipingIndex] = arr[swipingIndex] <= threshold ? snapOpen : snapClose;
    const days = [...this.data.days];
    if (swipingKey === 'breakfast') days[swipingDayIndex].swipeXBreakfast = arr;
    else if (swipingKey === 'lunch') days[swipingDayIndex].swipeXLunch = arr;
    else days[swipingDayIndex].swipeXDinner = arr;
    this.setData({ days, swipingDayIndex: -1, swipingKey: '', swipingIndex: -1, isSwiping: false });
  }
});
