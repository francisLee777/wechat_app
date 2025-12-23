// app.ts
import { initCurrentUser } from './utils/db';

App<IAppOption>({
  globalData: {},
  onLaunch() {
    // 初始化当前用户（模拟登录）
    const user = initCurrentUser();
    console.log('App Launch, Current User:', user);
  },
})
