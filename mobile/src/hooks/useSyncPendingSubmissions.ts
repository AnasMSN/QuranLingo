import { useEffect, useRef } from 'react';
import NetInfo from '@react-native-community/netinfo';
import { isAxiosError } from 'axios';

import { api } from '../api/client';
import { listPendingSubmissions, removePendingSubmission } from '../db/submissionQueue';
import { useAuthStore } from '../store/authStore';

// Replays lesson submissions that were queued while offline (see
// LessonScreen's submit flow), as soon as connectivity returns. The backend's
// idempotency_key makes every replay safe even if a submission actually made
// it through right before the connection dropped.
export function useSyncPendingSubmissions() {
  const applyUserPatch = useAuthStore((s) => s.applyUserPatch);
  const status = useAuthStore((s) => s.status);
  const flushing = useRef(false);

  useEffect(() => {
    const flush = async () => {
      if (flushing.current || status !== 'signed-in') return;
      flushing.current = true;
      try {
        const pending = await listPendingSubmissions();
        for (const submission of pending) {
          try {
            const result = await api.submitLesson(
              submission.lessonId,
              submission.idempotencyKey,
              submission.answers,
            );
            applyUserPatch({
              total_xp: result.total_xp,
              hearts: result.hearts_remaining,
              current_streak: result.current_streak,
              longest_streak: result.longest_streak,
            });
            await removePendingSubmission(submission.id);
          } catch (err) {
            const isNetworkError = isAxiosError(err) && !err.response;
            if (!isNetworkError) {
              // Server definitively rejected it (e.g. lesson removed) — no
              // amount of retrying will change that outcome.
              await removePendingSubmission(submission.id);
            } else {
              // Still offline / request timed out — stop and retry on the
              // next connectivity change rather than burning through retries.
              break;
            }
          }
        }
      } finally {
        flushing.current = false;
      }
    };

    const unsubscribe = NetInfo.addEventListener((state) => {
      if (state.isConnected) void flush();
    });

    void flush();
    return unsubscribe;
  }, [applyUserPatch, status]);
}
