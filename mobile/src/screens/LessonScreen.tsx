import { useMemo, useRef, useState } from 'react';
import { ActivityIndicator, Alert, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useQueryClient } from '@tanstack/react-query';
import { isAxiosError } from 'axios';

import { DuoButton } from '../components/DuoButton';
import { api, apiErrorMessage } from '../api/client';
import { useLessonDetail } from '../hooks/useCourse';
import { enqueueSubmission } from '../db/submissionQueue';
import { useAuthStore } from '../store/authStore';
import { colors, radii } from '../theme/colors';
import { generateId } from '../utils/id';
import type { AnswerInput } from '../types/api';
import type { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'Lesson'>;

export function LessonScreen({ route, navigation }: Props) {
  const { lessonId, lessonTitle } = route.params;
  const { data: lesson, isLoading, error } = useLessonDetail(lessonId);
  const applyUserPatch = useAuthStore((s) => s.applyUserPatch);
  const queryClient = useQueryClient();

  // Stable for the whole attempt so a retry (manual or offline-queue replay)
  // reuses the same idempotency_key and can never double-award XP.
  const idempotencyKey = useRef(generateId('submit')).current;

  const [index, setIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, AnswerInput>>({});
  const [textDraft, setTextDraft] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const exercise = lesson?.exercises[index];
  const total = lesson?.exercises.length ?? 0;
  const currentAnswer = exercise ? answers[exercise.id] : undefined;

  const canContinue = useMemo(() => {
    if (!exercise) return false;
    if (exercise.type === 'multiple_choice') return !!currentAnswer?.option_id;
    return textDraft.trim().length > 0;
  }, [exercise, currentAnswer, textDraft]);

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color={colors.green} />
      </View>
    );
  }

  if (error || !lesson || total === 0) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>Couldn&apos;t load this lesson.</Text>
        <DuoButton label="Go back" variant="outline" onPress={() => navigation.goBack()} style={{ marginTop: 16 }} />
      </View>
    );
  }

  const selectOption = (optionId: string) => {
    if (!exercise) return;
    setAnswers((prev) => ({ ...prev, [exercise.id]: { exercise_id: exercise.id, option_id: optionId } }));
  };

  const onContinue = async () => {
    if (!exercise) return;

    let nextAnswers = answers;
    if (exercise.type === 'translate') {
      nextAnswers = { ...answers, [exercise.id]: { exercise_id: exercise.id, text_answer: textDraft.trim() } };
      setAnswers(nextAnswers);
    }

    if (index + 1 < total) {
      setIndex(index + 1);
      setTextDraft('');
      return;
    }

    await finishLesson(nextAnswers);
  };

  const finishLesson = async (finalAnswers: Record<string, AnswerInput>) => {
    setSubmitting(true);
    const orderedAnswers = lesson.exercises.map((e) => finalAnswers[e.id] ?? { exercise_id: e.id });

    try {
      const result = await api.submitLesson(lessonId, idempotencyKey, orderedAnswers);
      applyUserPatch({
        total_xp: result.total_xp,
        hearts: result.hearts_remaining,
        current_streak: result.current_streak,
        longest_streak: result.longest_streak,
      });
      await queryClient.invalidateQueries({ queryKey: ['course'] });
      navigation.replace('LessonResults', { status: 'graded', result, lessonTitle });
    } catch (err) {
      const isNetworkError = isAxiosError(err) && !err.response;
      if (isNetworkError) {
        await enqueueSubmission({
          id: generateId('queued'),
          lessonId,
          lessonTitle,
          idempotencyKey,
          answers: orderedAnswers,
          createdAt: Date.now(),
        });
        navigation.replace('LessonResults', { status: 'queued', lessonTitle });
        return;
      }
      Alert.alert('Couldn\'t submit lesson', apiErrorMessage(err), [{ text: 'OK', onPress: () => navigation.goBack() }]);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.topBar}>
        <Pressable onPress={() => navigation.goBack()} hitSlop={12}>
          <Text style={styles.close}>✕</Text>
        </Pressable>
        <View style={styles.progressTrack}>
          <View style={[styles.progressFill, { width: `${((index + 1) / total) * 100}%` }]} />
        </View>
        <Text style={styles.heartIcon}>❤️</Text>
      </View>

      <View style={styles.body}>
        <Text style={styles.prompt}>{exercise!.prompt}</Text>
        {exercise!.arabic_text ? <Text style={styles.arabic}>{exercise!.arabic_text}</Text> : null}

        {exercise!.type === 'multiple_choice' ? (
          <View style={styles.options}>
            {exercise!.options?.map((opt) => {
              const selected = currentAnswer?.option_id === opt.id;
              return (
                <Pressable
                  key={opt.id}
                  onPress={() => selectOption(opt.id)}
                  style={[styles.option, selected && styles.optionSelected]}
                >
                  <Text style={[styles.optionText, selected && styles.optionTextSelected]}>{opt.option_text}</Text>
                </Pressable>
              );
            })}
          </View>
        ) : (
          <TextInput
            style={styles.textInput}
            placeholder="Type your answer"
            value={textDraft}
            onChangeText={setTextDraft}
            autoCapitalize="none"
            autoCorrect={false}
          />
        )}
      </View>

      <DuoButton
        label={index + 1 < total ? 'Continue' : 'Finish'}
        onPress={onContinue}
        disabled={!canContinue}
        loading={submitting}
        style={styles.continueButton}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg, paddingTop: 56, paddingHorizontal: 20 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24, gap: 8, backgroundColor: colors.bg },
  errorText: { fontSize: 16, fontWeight: '700', color: colors.text },
  topBar: { flexDirection: 'row', alignItems: 'center', gap: 12, marginBottom: 24 },
  close: { fontSize: 22, color: colors.textMuted, fontWeight: '700' },
  progressTrack: { flex: 1, height: 14, backgroundColor: colors.bgMuted, borderRadius: radii.pill, overflow: 'hidden' },
  progressFill: { height: '100%', backgroundColor: colors.green, borderRadius: radii.pill },
  heartIcon: { fontSize: 18 },
  body: { flex: 1 },
  prompt: { fontSize: 20, fontWeight: '800', color: colors.text, marginBottom: 12 },
  arabic: { fontSize: 40, textAlign: 'center', marginVertical: 24, color: colors.text },
  options: { gap: 12, marginTop: 12 },
  option: {
    borderWidth: 2,
    borderColor: colors.border,
    borderRadius: radii.lg,
    paddingVertical: 16,
    paddingHorizontal: 20,
  },
  optionSelected: { borderColor: colors.blue, backgroundColor: '#DDF4FF' },
  optionText: { fontSize: 16, fontWeight: '600', color: colors.text },
  optionTextSelected: { color: colors.blueDark },
  textInput: {
    borderWidth: 2,
    borderColor: colors.border,
    borderRadius: radii.lg,
    paddingHorizontal: 16,
    height: 56,
    fontSize: 18,
    marginTop: 12,
    color: colors.text,
  },
  continueButton: { marginBottom: 24 },
});
