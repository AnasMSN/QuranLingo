import { createAudioPlayer, type AudioPlayer } from 'expo-audio';

// Only one word clip should ever play at a time (auto-play on a new question,
// or the replay button), so track the current player and release it before
// starting the next one.
let currentPlayer: AudioPlayer | null = null;

// Plays a question's pronunciation clip. A missing or unreachable URL is an
// expected state for content that has no recording yet -- this fails
// completely silently (no thrown error, no user-facing message) rather than
// treating it as a bug.
export function playAudioUrl(url?: string | null) {
  if (!url) return;

  try {
    currentPlayer?.remove();
  } catch {
    // ignore -- player may already be released
  }
  currentPlayer = null;

  try {
    const player = createAudioPlayer(url);
    currentPlayer = player;
    player.play();
  } catch {
    currentPlayer = null;
  }
}
