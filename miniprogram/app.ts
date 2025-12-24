// app.ts
import { getCurrentUser, checkSession } from './api/auth';

App<IAppOption>({
  globalData: {},
  onLaunch() {
    // 检查 session 有效性
    checkSession().then(isValid => {
      if (isValid) {
        // session 有效，尝试获取本地用户信息
        const user = getCurrentUser();
        console.log('App Launch, Session Valid, User:', user);
      } else {
        console.log('App Launch, Session Invalid or Expired');
        // 可以在这里清除本地登录态，或者在 getCurrentUser 中处理
      }
    });
  },
})
