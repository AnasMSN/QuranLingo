import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native';

import { DuoButton } from '../components/DuoButton';
import { StatBadge } from '../components/StatBadge';
import { useAuthStore } from '../store/authStore';
import { colors, radii } from '../theme/colors';

export function ProfileScreen() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const onLogout = () => {
    Alert.alert('Log out?', 'You can always log back in.', [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Log out', style: 'destructive', onPress: () => void logout() },
    ]);
  };

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.avatar}>
        <Text style={styles.avatarText}>{(user?.display_name ?? '?').charAt(0).toUpperCase()}</Text>
      </View>
      <Text style={styles.name}>{user?.display_name}</Text>
      <Text style={styles.email}>{user?.email}</Text>

      <View style={styles.statsGrid}>
        <StatBadge icon="⭐" value={user?.total_xp ?? 0} color={colors.goldDark} />
        <StatBadge icon="❤️" value={user?.hearts ?? 0} color={colors.red} />
        <StatBadge icon="🔥" value={user?.current_streak ?? 0} color={colors.red} />
        <StatBadge icon="🏆" value={user?.longest_streak ?? 0} color={colors.blueDark} />
      </View>

      <View style={styles.card}>
        <Row label="Current streak" value={`${user?.current_streak ?? 0} days`} />
        <Row label="Longest streak" value={`${user?.longest_streak ?? 0} days`} />
        <Row label="Total XP" value={`${user?.total_xp ?? 0}`} />
        <Row label="Hearts" value={`${user?.hearts ?? 0} / 5`} />
      </View>

      <DuoButton label="Log out" variant="danger" onPress={onLogout} style={styles.logout} />
    </ScrollView>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, backgroundColor: colors.bg, alignItems: 'center', padding: 24, paddingTop: 72 },
  avatar: {
    width: 88,
    height: 88,
    borderRadius: radii.pill,
    backgroundColor: colors.green,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 12,
  },
  avatarText: { fontSize: 36, fontWeight: '800', color: '#FFFFFF' },
  name: { fontSize: 20, fontWeight: '800', color: colors.text },
  email: { fontSize: 14, color: colors.textMuted, marginBottom: 20 },
  statsGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 10, justifyContent: 'center', marginBottom: 24 },
  card: {
    width: '100%',
    backgroundColor: colors.bgMuted,
    borderRadius: radii.lg,
    padding: 16,
    marginBottom: 32,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  rowLabel: { fontSize: 14, color: colors.textMuted },
  rowValue: { fontSize: 14, fontWeight: '700', color: colors.text },
  logout: { width: '100%' },
});
