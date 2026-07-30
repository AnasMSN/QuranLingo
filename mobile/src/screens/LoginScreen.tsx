import { useState } from 'react';
import { KeyboardAvoidingView, Platform, StyleSheet, Text, TextInput, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { DuoButton } from '../components/DuoButton';
import { useAuthStore } from '../store/authStore';
import { colors, radii } from '../theme/colors';
import type { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'Login'>;

export function LoginScreen({ navigation }: Props) {
  const login = useAuthStore((s) => s.login);
  const error = useAuthStore((s) => s.error);
  const clearError = useAuthStore((s) => s.clearError);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    clearError();
    setSubmitting(true);
    try {
      await login(email.trim(), password);
    } catch {
      // error surfaced via the store
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <Text style={styles.logo}>🕌</Text>
      <Text style={styles.title}>QuranLingo</Text>
      <Text style={styles.subtitle}>Learn Arabic, one lesson at a time.</Text>

      <View style={styles.form}>
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
          placeholder="Password"
          secureTextEntry
          value={password}
          onChangeText={setPassword}
        />
        {error ? <Text style={styles.error}>{error}</Text> : null}

        <DuoButton
          label="Log in"
          onPress={onSubmit}
          disabled={!email || !password}
          loading={submitting}
          style={styles.button}
        />
        <DuoButton
          label="Create an account"
          variant="outline"
          onPress={() => navigation.navigate('Register')}
          style={styles.button}
        />
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.bg,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
  },
  logo: { fontSize: 64, marginBottom: 8 },
  title: { fontSize: 28, fontWeight: '800', color: colors.text },
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
