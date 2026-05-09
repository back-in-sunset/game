import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, ScrollView } from "react-native";
import { useAuthStore } from "../store/authStore";
import { serviceConfig } from "../services/config";

type Props = { onBack: () => void };

export function SettingsScreen({ onBack }: Props) {
  const { baseUrl, login, logout, isLoggedIn } = useAuthStore();
  const [inputUrl, setInputUrl] = useState(baseUrl || "http://localhost:8080");
  const [inputToken, setInputToken] = useState("");

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={onBack}>
          <Text style={styles.backBtn}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.title}>设置</Text>
      </View>

      <View style={styles.card}>
        <Text style={styles.cardTitle}>服务配置</Text>
        {(["imWsUrl", "vdaGrpcUrl", "livekitUrl"] as const).map((k) => (
          <View key={k} style={styles.row}>
            <Text style={styles.label}>{k}</Text>
            <Text style={styles.value}>{serviceConfig[k] || "unset"}</Text>
          </View>
        ))}
      </View>

      <View style={styles.card}>
        <Text style={styles.cardTitle}>登录</Text>
        {isLoggedIn() ? (
          <>
            <Text style={styles.loggedIn}>已登录: {baseUrl}</Text>
            <TouchableOpacity style={styles.logoutBtn} onPress={logout}>
              <Text style={styles.logoutText}>退出登录</Text>
            </TouchableOpacity>
          </>
        ) : (
          <>
            <Text style={styles.label}>API 地址</Text>
            <TextInput style={styles.input} value={inputUrl} onChangeText={setInputUrl} placeholder="http://localhost:8080" autoCapitalize="none" />
            <Text style={styles.label}>JWT Token</Text>
            <TextInput style={styles.input} value={inputToken} onChangeText={setInputToken} placeholder="输入 Token" secureTextEntry autoCapitalize="none" />
            <TouchableOpacity style={styles.loginBtn} onPress={() => login(inputUrl.trim(), inputToken.trim())}>
              <Text style={styles.loginText}>登录</Text>
            </TouchableOpacity>
          </>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  header: { flexDirection: "row", alignItems: "center", padding: 16, paddingTop: 60, backgroundColor: "#fff", borderBottomWidth: 1, borderColor: "#e2e8f0" },
  backBtn: { fontSize: 15, color: "#2f6bff", marginRight: 12 },
  title: { fontSize: 18, fontWeight: "600", color: "#0f172a" },
  card: { backgroundColor: "#fff", margin: 16, padding: 16, borderRadius: 12 },
  cardTitle: { fontSize: 16, fontWeight: "600", color: "#0f172a", marginBottom: 12 },
  row: { flexDirection: "row", justifyContent: "space-between", paddingVertical: 8 },
  label: { fontSize: 13, color: "#64748b", marginBottom: 4 },
  value: { fontSize: 13, color: "#0f172a", fontWeight: "500" },
  input: { backgroundColor: "#f1f5f9", borderRadius: 8, padding: 10, fontSize: 14, marginBottom: 12 },
  loginBtn: { backgroundColor: "#2f6bff", borderRadius: 10, paddingVertical: 12, alignItems: "center" as const, marginTop: 8 },
  loginText: { color: "#fff", fontWeight: "600", fontSize: 15 },
  loggedIn: { fontSize: 14, color: "#16a34a", marginBottom: 8 },
  logoutBtn: { backgroundColor: "#fee2e2", borderRadius: 8, paddingVertical: 10, alignItems: "center" as const },
  logoutText: { color: "#dc2626", fontWeight: "600", fontSize: 14 },
});
