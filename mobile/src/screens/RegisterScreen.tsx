import { useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { DuoButton } from '../components/DuoButton';
import { useAuthStore } from '../store/authStore';
import { colors, radii } from '../theme/colors';
import type { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'Register'>;

export function RegisterScreen({ navigation }: Props) {
  const register = useAuthStore((s) => s.register);
  const error = useAuthStore((s) => s.error);
  const clearError = useAuthStore((s) => s.clearError);
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    clearError();
    setSubmitting(true);
    try {
      await register(email.trim(), password, displayName.trim());
    } catch {
      // error surfaced via the store
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView style={styles.container} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView contentContainerStyle={styles.scroll} keyboardShouldPersistTaps="handled">
        <Text style={styles.title}>Create your account</Text>
        <Text style={styles.subtitle}>Start your Arabic learning streak today.</Text>

        <View style={styles.form}>
          <TextInput
            style={styles.input}
            placeholder="Display name"
            autoCapitalize="words"
            value={displayName}
            onChangeText={setDisplayName}
          />
          <TextInput
            style={styles.input}
            placeholder="Email"
            autoCapitalize="none"
            keyboardType="email-address"
            value={email}
            onChangeText={setEmail}
          />
          <TextInput
            style={styles.input}
            placeholder="Password (min. 8 characters)"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />
          {error ? <Text style={styles.error}>{error}</Text> : null}

          <DuoButton
            label="Sign up"
            onPress={onSubmit}
            disabled={!email || password.length < 8 || !displayName}
            loading={submitting}
            style={styles.button}
          />
          <DuoButton label="Back to login" variant="outline" onPress={() => navigation.goBack()} style={styles.button} />
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  scroll: { flexGrow: 1, justifyContent: 'center', padding: 24 },
  title: { fontSize: 26, fontWeight: '800', color: colors.text },
  subtitle: { fontSize: 15, color: colors.textMuted, marginTop: 4, marginBottom: 32 },
  form: { width: '100%' },
  input: {
    height: 52,
    borderWidth: 2,
    borderColor: colors.border,
    borderRadius: radii.lg,
    paddingHorizontal: 16,
    fontSize: 16,
    marginBottom: 12,
    color: colors.text,
  },
  error: { color: colors.red, marginBottom: 12, fontWeight: '600' },
  button: { marginTop: 8, width: '100%' },
});
