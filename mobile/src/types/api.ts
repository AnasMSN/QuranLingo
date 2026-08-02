// Types mirror the Go backend's JSON contracts exactly (see backend/internal/service
// and backend/internal/handler). Keep in sync by hand until an OpenAPI spec exists.

export interface UserResponse {
  id: string;
  email: string;
  display_name: string;
  total_xp: number;
  hearts: number;
  current_streak: number;
  longest_streak: number;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

export interface AuthResponse {
  user: UserResponse;
  tokens: TokenPair;
}

export type LessonStatus = 'locked' | 'unlocked' | 'completed';

export interface LessonNode {
  id: string;
  title: string;
  position: number;
  xp_reward: number;
  status: LessonStatus;
}

export interface SkillNode {
  id: string;
  code: string;
  title: string;
  description: string;
  icon: string;
  position: number;
  status: LessonStatus;
  lessons: LessonNode[];
}

export interface CourseTree {
  id: string;
  code: string;
  title: string;
  description: string;
  skills: SkillNode[];
}

// The app only ever supports multiple-choice questions.
export type ExerciseType = 'multiple_choice';

export interface ExerciseOptionDTO {
  id: string;
  option_text: string;
}

export interface ExerciseDTO {
  id: string;
  type: ExerciseType;
  prompt: string;
  arabic_text?: string;
  audio_url?: string;
  options?: ExerciseOptionDTO[];
}

export interface LessonDetail {
  id: string;
  title: string;
  xp_reward: number;
  exercises: ExerciseDTO[];
}

export interface AnswerInput {
  exercise_id: string;
  option_id?: string;
}

export interface ExerciseResult {
  exercise_id: string;
  correct: boolean;
  correct_answer: string;
}

export interface SubmitResult {
  already_submitted: boolean;
  score: number;
  xp_earned: number;
  total_xp: number;
  hearts_remaining: number;
  current_streak: number;
  longest_streak: number;
  results?: ExerciseResult[];
}

export interface LeaderboardEntry {
  UserID: string;
  DisplayName: string;
  WeeklyXP: number;
}

export interface LeaderboardResponse {
  entries: LeaderboardEntry[];
}

export interface ApiErrorBody {
  error: string;
}
