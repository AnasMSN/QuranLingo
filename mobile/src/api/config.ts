import { Platform } from 'react-native';

// The backend (see ../../../backend) listens on APP_PORT (default 8080) with no
// path prefix. On a physical device or a Wi-Fi-connected simulator, "localhost"
// refers to the device itself, not your dev machine — set EXPO_PUBLIC_API_URL to
// your machine's LAN IP (e.g. http://192.168.1.20:8080) in that case.
const ANDROID_EMULATOR_HOST = 'http://10.0.2.2:8080';
const DEFAULT_HOST = 'http://localhost:8080';

export const API_URL =
  process.env.EXPO_PUBLIC_API_URL ??
  (Platform.OS === 'android' ? ANDROID_EMULATOR_HOST : DEFAULT_HOST);
