# RTSP Stream Setup

## Konfigurasi
Konfigurasikan RTSP stream di file `config.json`:

```json
"streams": {
    "reowhite": { // UUID stream
      "VOD": false,
      "disableAudio": true,
      "debug": false,
      "url": "rtsp://admin:admin@192.168.1.11:1935" // URL RTSP lokal
    }
  }
```

## Cara Menjalankan

### 1. Instalasi Dependensi
```sh
npm install
```

### 2. Build Aplikasi
- **Untuk Linux:**
  ```sh
  npm run build
  ```
- **Untuk Windows:**
  ```sh
  npm run build-wd
  ```

### 3. Jalankan Aplikasi
```sh
npm start
```

## Akses RTSP Stream
- **Lokal:**
  ```
http://localhost:8000/webrtc.html```

- **Akses Publik:**
  Gunakan **ngrok** untuk membuat akses publik:
  ```sh
  ngrok http 8002
  ```
  Setelah itu, gunakan URL yang diberikan oleh ngrok untuk mengakses stream dari luar jaringan lokal:
  ```
  https://{ngrok_url}/stream/webrtc/{uuid}
  ```

---

### Catatan
- Pastikan URL RTSP yang digunakan benar dan dapat diakses dari server.
- Gunakan mode **debug** (`"debug": true`) dalam `config.json` jika mengalami kendala.
- Pastikan firewall tidak memblokir port yang diperlukan untuk streaming.