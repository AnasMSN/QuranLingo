import { create } from 'zustand';

import { api, apiErrorMessage, hydrateAuth, setOnAuthFailure } from '../api/client';
import type { UserResponse } from '../types/api';

interface AuthState {
  status: 'loading' | 'signed-out' | 'signed-in';
  user: UserResponse | null;
  error: string | null;
  bootstrap: () => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  /** Optimistically merge fields (e.g. after a lesson submit) without a round trip. */
  applyUserPatch: (patch: Partial<UserResponse>) => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  status: 'loading',
  user: null,
  error: null,

  bootstrap: async () => {
    const user = await hydrateAuth();
    set({ status: user ? 'signed-in' : 'signed-out', user });
  },

  login: async (email, password) => {
    set({ error: null });
    try {
      const res = await api.login(email, password);
      set({ status: 'signed-in', user: res.user });
    } catch (err) {
      set({ error: apiErrorMessage(err, 'Login failed') });
      throw err;
    }
  },

  register: async (email, password, displayName) => {
    set({ error: null });
    try {
      const res = await api.register(email, password, displayName);
      set({ status: 'signed-in', user: res.user });
    } catch (err) {
      set({ error: apiErrorMessage(err, 'Registration failed') });
      throw err;
    }
  },

  logout: async () => {
    await api.logout();
    set({ status: 'signed-out', user: null });
  },

  refreshUser: async () => {
    if (get().status !== 'signed-in') return;
    const user = await api.me();
    set({ user });
  },

  applyUserPatch: (patch) => {
    const current = get().user;
    if (!current) return;
    set({ user: { ...current, ...patch } });
  },

  clearError: () => set({ error: null }),
}));

// Wire the API client's "refresh token is dead" callback to a real logout so
// a 401 the client can't recover from always lands the user on the auth stack.
setOnAuthFailure(() => {
  useAuthStore.setState({ status: 'signed-out', user: null });
});
