export const compressImage = (src: string, quality: number = 0.8, maxSide: number = 2048): Promise<string> => {
  return new Promise((resolve) => {
    wx.getImageInfo({
      src,
      success: (res) => {
        const w = res.width;
        const h = res.height;
        const path = res.path;
        let tw = w;
        let th = h;
        const fs = wx.getFileSystemManager();
        const MAX_SIZE = 300 * 1024;

        const makeCanvas = (width: number, height: number) => {
          const canvas = wx.createOffscreenCanvas({ type: '2d', width, height });
          const ctx = canvas.getContext('2d');
          const img = canvas.createImage();
          return { canvas, ctx, img };
        };

        const exportOnce = (canvas: any, width: number, height: number, q: number) => new Promise<string>((res2, rej2) => {
          wx.canvasToTempFilePath({
            canvas,
            fileType: 'jpg',
            quality: q,
            destWidth: width,
            destHeight: height,
            success: (r) => res2(r.tempFilePath),
            fail: (e) => rej2(e)
          });
        });

        const getSize = (p: string) => new Promise<number>((res3) => {
          fs.getFileInfo({
            filePath: p,
            success: (r) => res3(r.size),
            fail: () => res3(0)
          });
        });

        const run = async () => {
          if (Math.max(w, h) > maxSide) {
            const scale = maxSide / Math.max(w, h);
            tw = Math.max(1, Math.floor(w * scale));
            th = Math.max(1, Math.floor(h * scale));
          }

          let q = quality;
          let attempts = 0;
          let scaleAttempts = 0;

          while (true) {
            const { canvas, ctx, img } = makeCanvas(tw, th);
            await new Promise<void>((r) => {
              img.onload = () => { ctx.clearRect(0, 0, tw, th); ctx.drawImage(img, 0, 0, tw, th); r(); };
              img.onerror = () => { r(); };
              img.src = path;
            });

            let temp: string;
            try {
              temp = await exportOnce(canvas, tw, th, q);
            } catch {
              resolve(path);
              return;
            }

            const size = await getSize(temp);
            if (size > 0 && size <= MAX_SIZE) { resolve(temp); return; }

            if (q > 0.3 && attempts < 6) { q -= 0.1; attempts++; continue; }

            if (scaleAttempts < 8) {
              const factor = 0.85;
              tw = Math.max(1, Math.floor(tw * factor));
              th = Math.max(1, Math.floor(th * factor));
              q = quality;
              scaleAttempts++;
              continue;
            }

            resolve(temp);
            return;
          }
        };

        run();
      },
      fail: () => { resolve(src); }
    });
  });
};
