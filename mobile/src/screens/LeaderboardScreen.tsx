import { ActivityIndicator, FlatList, StyleSheet, Text, View } from 'react-native';
import { useQuery } from '@tanstack/react-query';

import { api } from '../api/client';
import { useAuthStore } from '../store/authStore';
import { colors, radii } from '../theme/colors';
import type { LeaderboardEntry } from '../types/api';

const MEDALS = ['🥇', '🥈', '🥉'];

export function LeaderboardScreen() {
  const currentUserId = useAuthStore((s) => s.user?.id);
  const { data, isLoading, error, refetch, isRefetching } = useQuery({
    queryKey: ['leaderboard', 'weekly'],
    queryFn: () => api.leaderboardWeekly(),
    staleTime: 30_000,
  });

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color={colors.green} />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>Couldn&apos;t load the leaderboard.</Text>
      </View>
    );
  }

  const renderItem = ({ item, index }: { item: LeaderboardEntry; index: number }) => {
    const isMe = item.UserID === currentUserId;
    return (
      <View style={[styles.row, isMe && styles.rowMe]}>
        <Text style={styles.rank}>{MEDALS[index] ?? `#${index + 1}`}</Text>
        <Text style={[styles.name, isMe && styles.nameMe]} numberOfLines={1}>
          {item.DisplayName}
          {isMe ? ' (You)' : ''}
        </Text>
        <Text style={styles.xp}>{item.WeeklyXP} XP</Text>
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>🏆 Weekly Leaderboard</Text>
      </View>
      <FlatList
        data={data?.entries ?? []}
        keyExtractor={(item) => item.UserID}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
        onRefresh={refetch}
        refreshing={isRefetching}
        ListEmptyComponent={
          <Text style={styles.empty}>No XP earned this week yet — complete a lesson to get on the board!</Text>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bg },
  errorText: { fontSize: 16, fontWeight: '700', color: colors.text },
  header: { paddingTop: 56, paddingHorizontal: 20, paddingBottom: 16 },
  title: { fontSize: 22, fontWeight: '800', color: colors.text },
  list: { paddingHorizontal: 20, paddingBottom: 24 },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingVertical: 14,
    paddingHorizontal: 12,
    borderRadius: radii.md,
    marginBottom: 8,
    backgroundColor: colors.bgMuted,
  },
  rowMe: { backgroundColor: '#DDF4FF' },
  rank: { fontSize: 18, width: 32, textAlign: 'center' },
  name: { flex: 1, fontSize: 15, fontWeight: '700', color: colors.text },
  nameMe: { color: colors.blueDark },
  xp: { fontSize: 14, fontWeight: '800', color: colors.goldDark },
  empty: { textAlign: 'center', color: colors.textMuted, marginTop: 40, paddingHorizontal: 20 },
});
