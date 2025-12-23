// 模拟后端数据库服务
// 使用 wx.setStorageSync 和 wx.getStorageSync 进行本地数据持久化
// 注意：真实项目中应使用云开发数据库或后端 API

// ================= 类型定义 =================

// 用户角色：家长 或 成员
export type UserRole = 'admin' | 'member';

// 用户信息结构
export interface User {
  id: string;          // 用户唯一标识
  nickName: string;    // 用户昵称
  avatarUrl: string;   // 用户头像 URL
  familyId: string | null; // 所属家庭 ID，为 null 表示未加入任何家庭
  role: UserRole;      // 在家庭中的角色
}

// 家庭信息结构
export interface Family {
  id: string;          // 家庭唯一标识
  name: string;        // 家庭名称
  ownerId: string;     // 创建者（家长）的用户 ID
  createTime: number;  // 创建时间戳
}

// 食谱信息结构
export interface Recipe {
  id: string;          // 食谱唯一标识
  familyId: string;    // 所属家庭 ID
  name: string;        // 食谱名称
  images: string[];    // 成品照片路径列表（新版：多图）
  image?: string;      // 兼容旧版：单图（已废弃）
  content: string;     // 食谱具体内容
  createTime: number;  // 创建时间戳
  updateTime: number;  // 更新时间戳
  sortOrder: number;   // 排序字段（值越大越靠前）
}

// 留言信息结构（新设计：扁平化）
export interface Message {
  id: string;          // 留言唯一标识
  familyId: string;    // 所属家庭 ID
  userId: string;      // 留言者用户 ID
  userName: string;    // 留言者昵称（冗余存储方便展示）
  userAvatar: string;  // 留言者头像（冗余存储方便展示）
  content: string;     // 留言内容
  createTime: number;  // 创建时间戳
  // 如果是回复，则有 parentId
  parentId: string | null; // 父留言 ID，如果是根留言则为 null
  replyToUserId?: string; // 回复给谁（冗余）
  replyToUserName?: string; // 回复给谁的昵称（冗余）
}

// ================= 常量定义 =================

const KEY_USERS = 'DB_USERS';       // 存储所有用户的 Key
const KEY_FAMILIES = 'DB_FAMILIES'; // 存储所有家庭的 Key
const KEY_RECIPES = 'DB_RECIPES';   // 存储所有食谱的 Key
const KEY_MESSAGES = 'DB_MESSAGES'; // 存储所有留言的 Key
const KEY_CURRENT_USER_ID = 'CURRENT_USER_ID'; // 当前登录用户的 ID

// ================= 辅助函数 =================

// 生成唯一 ID (简单的随机字符串)
function generateId(): string {
  return Math.random().toString(36).substr(2, 9) + Date.now().toString(36);
}

// 获取所有数据（通用）
function getData<T>(key: string): T {
  return (wx.getStorageSync(key) || (key === KEY_RECIPES || key === KEY_MESSAGES ? [] : {})) as T;
}

// 保存数据（通用）
function saveData(key: string, data: any) {
  wx.setStorageSync(key, data);
}

// ================= 用户相关 =================

// 初始化或获取当前用户
// 模拟登录过程，如果本地没有用户ID，则创建一个新用户
export function initCurrentUser(): User {
  let userId = wx.getStorageSync(KEY_CURRENT_USER_ID);
  const users = getData<{[key: string]: User}>(KEY_USERS);

  if (!userId || !users[userId]) {
    // 创建新用户
    userId = generateId();
    const newUser: User = {
      id: userId,
      nickName: `用户${userId.substr(0, 4)}`, // 默认昵称
      avatarUrl: '', // 默认头像为空
      familyId: null,
      role: 'member' // 默认为成员，创建家庭时会改为 admin
    };
    users[userId] = newUser;
    saveData(KEY_USERS, users);
    wx.setStorageSync(KEY_CURRENT_USER_ID, userId);
    console.log('创建新用户:', newUser);
    return newUser;
  }
  
  return users[userId];
}

// 获取当前用户信息
export function getCurrentUser(): User {
  return initCurrentUser();
}

// 更新用户信息
export function updateUser(user: User) {
  const users = getData<{[key: string]: User}>(KEY_USERS);
  users[user.id] = user;
  saveData(KEY_USERS, users);
}

// 根据ID获取用户
export function getUserById(id: string): User | null {
  const users = getData<{[key: string]: User}>(KEY_USERS);
  return users[id] || null;
}

// ================= 家庭相关 =================

// 创建家庭
export function createFamily(name: string): Family {
  const user = getCurrentUser();
  if (user.familyId) {
    throw new Error('您已加入家庭，无法创建新家庭');
  }

  const familyId = generateId();
  const newFamily: Family = {
    id: familyId,
    name: name,
    ownerId: user.id,
    createTime: Date.now()
  };

  // 保存家庭数据
  const families = getData<{[key: string]: Family}>(KEY_FAMILIES);
  families[familyId] = newFamily;
  saveData(KEY_FAMILIES, families);

  // 更新用户状态
  user.familyId = familyId;
  user.role = 'admin'; // 创建者自动成为家长
  updateUser(user);

  return newFamily;
}

// 加入家庭
export function joinFamily(familyId: string): Family {
  const user = getCurrentUser();
  if (user.familyId) {
    throw new Error('您已加入家庭，无法加入其他家庭');
  }

  const families = getData<{[key: string]: Family}>(KEY_FAMILIES);
  const family = families[familyId];
  if (!family) {
    throw new Error('家庭不存在，请检查ID是否正确');
  }

  // 检查家庭成员数量限制 (最多4人)
  const members = getFamilyMembers(familyId);
  if (members.length >= 4) {
    throw new Error('该家庭成员已满4人，无法加入');
  }

  // 更新用户状态
  user.familyId = familyId;
  user.role = 'member'; // 加入者默认为成员
  updateUser(user);

  return family;
}

// 根据ID获取家庭
export function getFamilyById(id: string): Family | undefined {
  const families = getData<{[key: string]: Family}>(KEY_FAMILIES);
  return families[id];
}

// 获取家庭成员列表
export function getFamilyMembers(familyId: string): User[] {
  const users = getData<{[key: string]: User}>(KEY_USERS);
  // 遍历所有用户，找到 familyId 匹配的用户
  return Object.values(users).filter(u => u.familyId === familyId);
}

// 移除成员（家长功能）
export function removeMember(targetUserId: string) {
  const currentUser = getCurrentUser();
  const targetUser = getUserById(targetUserId);

  if (!targetUser) throw new Error('用户不存在');
  if (currentUser.role !== 'admin') throw new Error('只有家长可以移除成员');
  if (currentUser.familyId !== targetUser.familyId) throw new Error('该成员不在您的家庭中');
  if (currentUser.id === targetUser.id) throw new Error('不能移除自己');

  // 更新目标用户状态：清除家庭ID和角色
  targetUser.familyId = null;
  targetUser.role = 'member'; // 重置为默认
  updateUser(targetUser);
}

// 退出家庭（成员功能）
export function quitFamily() {
  const user = getCurrentUser();
  if (!user.familyId) throw new Error('您还没有加入家庭');
  if (user.role === 'admin') throw new Error('家长不能退出家庭，只能解散家庭（暂未实现解散）');

  user.familyId = null;
  updateUser(user);
}

// ================= 食谱相关 =================

// 数据兼容处理：旧数据 image: string -> images: string[]
// 同时处理 sortOrder 缺失问题
function normalizeRecipe(r: Recipe): Recipe {
  if (!r.images && r.image) {
    r.images = [r.image];
  }
  if (!r.images) {
    r.images = [];
  }
  if (typeof r.sortOrder !== 'number') {
    r.sortOrder = r.createTime;
  }
  return r;
}

// 添加食谱
export function addRecipe(name: string, images: string[], content: string): Recipe {
  const user = getCurrentUser();
  if (!user.familyId) throw new Error('请先加入家庭');
  if (user.role !== 'admin') throw new Error('只有家长可以创建食谱');

  const now = Date.now();
  const newRecipe: Recipe = {
    id: generateId(),
    familyId: user.familyId,
    name,
    images,
    content,
    createTime: now,
    updateTime: now,
    sortOrder: now // 默认排序值为创建时间，越新越靠前
  };

  const recipes = getData<Recipe[]>(KEY_RECIPES);
  recipes.push(newRecipe);
  saveData(KEY_RECIPES, recipes);

  return newRecipe;
}

// 获取家庭食谱列表（支持搜索）
export function getRecipes(keyword: string = ''): Recipe[] {
  const user = getCurrentUser();
  if (!user.familyId) return [];

  const recipes = getData<Recipe[]>(KEY_RECIPES);
  // 过滤当前家庭的食谱，并进行数据标准化
  let familyRecipes = recipes
    .filter(r => r.familyId === user.familyId)
    .map(normalizeRecipe);

  // 关键词搜索
  if (keyword) {
    familyRecipes = familyRecipes.filter(r => r.name.includes(keyword) || r.content.includes(keyword));
  }
  
  // 按 sortOrder 倒序排列 (值越大越靠前)
  return familyRecipes.sort((a, b) => b.sortOrder - a.sortOrder);
}

// 获取单个食谱详情
export function getRecipeById(id: string): Recipe | undefined {
  const recipes = getData<Recipe[]>(KEY_RECIPES);
  const recipe = recipes.find(r => r.id === id);
  return recipe ? normalizeRecipe(recipe) : undefined;
}

// 更新食谱
export function updateRecipe(id: string, name: string, images: string[], content: string) {
  const user = getCurrentUser();
  if (user.role !== 'admin') throw new Error('只有家长可以编辑食谱');

  const recipes = getData<Recipe[]>(KEY_RECIPES);
  const index = recipes.findIndex(r => r.id === id);
  
  if (index === -1) throw new Error('食谱不存在');
  if (recipes[index].familyId !== user.familyId) throw new Error('无权编辑此食谱');

  recipes[index].name = name;
  recipes[index].images = images;
  // 兼容：清空旧的 image 字段以防混淆
  delete recipes[index].image;
  recipes[index].content = content;
  recipes[index].updateTime = Date.now();
  
  saveData(KEY_RECIPES, recipes);
}

// 调整食谱排序 (上移/下移)
// direction: -1 (上移/前移，因为倒序所以是增加 sortOrder), 1 (下移/后移，减少 sortOrder)
// 由于 sortOrder 是时间戳级别的大整数，直接交换两者的 sortOrder 最简单
export function reorderRecipe(id: string, direction: 'up' | 'down') {
  const user = getCurrentUser();
  if (user.role !== 'admin') throw new Error('只有家长可以调整排序');

  // 获取并标准化当前列表（按顺序）
  const recipes = getData<Recipe[]>(KEY_RECIPES);
  const familyRecipes = recipes
    .filter(r => r.familyId === user.familyId)
    .map(normalizeRecipe)
    .sort((a, b) => b.sortOrder - a.sortOrder);

  const currentIndex = familyRecipes.findIndex(r => r.id === id);
  if (currentIndex === -1) return;

  const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;

  // 边界检查
  if (targetIndex < 0 || targetIndex >= familyRecipes.length) return;

  // 交换 sortOrder
  const currentRecipe = familyRecipes[currentIndex];
  const targetRecipe = familyRecipes[targetIndex];

  const tempOrder = currentRecipe.sortOrder;
  currentRecipe.sortOrder = targetRecipe.sortOrder;
  targetRecipe.sortOrder = tempOrder;

  // 如果 sortOrder 相等（极端情况），手动微调
  if (currentRecipe.sortOrder === targetRecipe.sortOrder) {
    if (direction === 'up') currentRecipe.sortOrder += 1;
    else targetRecipe.sortOrder += 1;
  }

  // 更新回总列表 (找到原始引用进行修改，或者通过 ID 查找更新)
  // 因为 familyRecipes 是 filter+map 出来的新对象，所以直接修改它们不会影响 recipes
  // 需要把更新后的 sortOrder 写回 recipes
  const realCurrent = recipes.find(r => r.id === currentRecipe.id);
  const realTarget = recipes.find(r => r.id === targetRecipe.id);
  
  if (realCurrent && realTarget) {
    realCurrent.sortOrder = currentRecipe.sortOrder;
    realTarget.sortOrder = targetRecipe.sortOrder;
    // 同时也把 images 字段可能做的迁移写回去
    if (currentRecipe.images) realCurrent.images = currentRecipe.images;
    if (targetRecipe.images) realTarget.images = targetRecipe.images;
  }

  saveData(KEY_RECIPES, recipes);
}

// 删除食谱
export function deleteRecipe(id: string) {
  const user = getCurrentUser();
  if (user.role !== 'admin') throw new Error('只有家长可以删除食谱');

  let recipes = getData<Recipe[]>(KEY_RECIPES);
  // 过滤掉要删除的那个
  const originalLength = recipes.length;
  recipes = recipes.filter(r => !(r.id === id && r.familyId === user.familyId));
  
  if (recipes.length === originalLength) {
     // 没有变化
  }

  saveData(KEY_RECIPES, recipes);
}

// 批量更新食谱（用于拖拽排序保存）
export function batchUpdateRecipes(updatedRecipes: Recipe[]) {
  const user = getCurrentUser();
  if (user.role !== 'admin') throw new Error('无权操作');

  let allRecipes = getData<Recipe[]>(KEY_RECIPES);
  
  updatedRecipes.forEach(ur => {
    const idx = allRecipes.findIndex(r => r.id === ur.id);
    if (idx !== -1 && allRecipes[idx].familyId === user.familyId) {
      allRecipes[idx].sortOrder = ur.sortOrder;
    }
  });

  saveData(KEY_RECIPES, allRecipes);
}


// ================= 留言板相关 =================

// 添加留言（支持回复）
export function addMessage(content: string, parentId: string | null = null, replyToUserId?: string, replyToUserName?: string): Message {
  const user = getCurrentUser();
  if (!user.familyId) throw new Error('请先加入家庭');

  const newMessage: Message = {
    id: generateId(),
    familyId: user.familyId,
    userId: user.id,
    userName: user.nickName,
    userAvatar: user.avatarUrl,
    content,
    createTime: Date.now(),
    parentId,
    replyToUserId,
    replyToUserName
  };

  const messages = getData<Message[]>(KEY_MESSAGES);
  messages.push(newMessage);
  saveData(KEY_MESSAGES, messages);

  return newMessage;
}

// 获取家庭留言列表（结构化）
// 返回根留言，每个根留言包含 replies 数组
export function getMessages(): any[] {
  const user = getCurrentUser();
  if (!user.familyId) return [];

  // 获取该家庭所有消息
  let messages = getData<any[]>(KEY_MESSAGES);
  
  // 简单的旧数据兼容清洗
  messages = messages.filter(m => {
    if (m.familyId !== user.familyId) return false;
    // 如果是旧数据（有 replies 数组的），需要拆解（这里简化处理，直接丢弃旧结构，或者假设已清洗）
    // 为了稳健，我们假设数据已经符合新结构，或者我们只过滤符合新结构的
    // 实际生产中需要写迁移脚本，这里为了演示，我们直接过滤出扁平化结构
    // 但考虑到之前的反复修改，我们做一个强制转换
    return true; 
  });

  // 1. 找出所有根留言 (parentId == null)
  const rootMessages = messages.filter(m => !m.parentId || m.parentId === null);
  
  // 2. 找出所有回复
  const replyMessages = messages.filter(m => m.parentId);

  // 3. 组装
  const result = rootMessages.map(root => {
    const replies = replyMessages
      .filter(r => r.parentId === root.id)
      .sort((a, b) => a.createTime - b.createTime); // 回复按时间正序
    
    return {
      ...root,
      replies
    };
  });

  // 根留言按时间倒序
  return result.sort((a, b) => b.createTime - a.createTime);
}

// 删除留言
export function deleteMessage(msgId: string) {
  const user = getCurrentUser();
  let messages = getData<Message[]>(KEY_MESSAGES);
  
  const msgIndex = messages.findIndex(m => m.id === msgId);
  if (msgIndex === -1) throw new Error('留言不存在');
  
  const msg = messages[msgIndex];
  
  // 权限校验：家长可以删除任意留言，成员只能删除自己的
  if (user.role !== 'admin' && msg.userId !== user.id) {
    throw new Error('您只能删除自己的留言');
  }

  // 如果是删除根留言，需要级联删除其所有回复
  if (!msg.parentId) {
    messages = messages.filter(m => m.id !== msgId && m.parentId !== msgId);
  } else {
    // 只是删除一条回复
    messages.splice(msgIndex, 1);
  }

  saveData(KEY_MESSAGES, messages);
}
