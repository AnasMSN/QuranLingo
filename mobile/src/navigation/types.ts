import type { SubmitResult } from '../types/api';

export type LessonResultsParams =
  | { lessonTitle: string; status: 'graded'; result: SubmitResult }
  | { lessonTitle: string; status: 'queued' };

export type RootStackParamList = {
  Login: undefined;
  Register: undefined;
  Main: undefined;
  Lesson: { lessonId: string; lessonTitle: string };
  LessonResults: LessonResultsParams;
};

export type MainTabParamList = {
  Learn: undefined;
  Leaderboard: undefined;
  Profile: undefined;
};

declare global {
  namespace ReactNavigation {
    // eslint-disable-next-line @typescript-eslint/no-empty-object-type -- required shape for React Navigation's global type augmentation
    interface RootParamList extends RootStackParamList {}
  }
}
