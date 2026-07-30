import { useEffect } from 'react';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { RootNavigator } from './src/navigation/RootNavigator';
import { useAuthStore } from './src/store/authStore';
import { useSyncPendingSubmissions } from './src/hooks/useSyncPendingSubmissions';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false },
  },
});

function AppInner() {
  const bootstrap = useAuthStore((s) => s.bootstrap);
  useSyncPendingSubmissions();

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  return (
    <>
      <RootNavigator />
      <StatusBar style="dark" />
    </>
  );
}

export default function App() {
  return (
    <SafeAreaProvider>
      <QueryClientProvider client={queryClient}>
        <AppInner />
      </QueryClientProvider>
    </SafeAreaProvider>
  );
}
