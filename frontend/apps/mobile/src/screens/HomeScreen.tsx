import { useEffect } from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { conversations } from "@game/shared";
import { useIMStore } from "../store/imStore";
import { serviceConfig } from "../services/config";

type Props = { navigate: (screen: string, params?: Record<string, unknown>) => void };

export function HomeScreen({ navigate }: Props) {
  const { status, error, connect, disconnect } = useIMStore();

  useEffect(() => {
    connect(serviceConfig.imWsUrl, {
      token: serviceConfig.imToken,
      domain: serviceConfig.imDomain,
      scope: {
        tenant_id: serviceConfig.imTenantId,
        project_id: serviceConfig.imProjectId,
        environment: serviceConfig.imEnvironment,
      },
    });
    return () => disconnect();
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>VDA + IM</Text>
      <View style={[styles.statusBadge, { backgroundColor: status === "connected" ? "#dcfce7" : "#fef3c7" }]}>
        <Text style={styles.statusText}>{status}</Text>
      </View>
      {error ? <Text style={styles.error}>{error}</Text> : null}

      <Text style={styles.sectionTitle}>会话</Text>
      {conversations.map((c) => (
        <TouchableOpacity
          key={c.id}
          style={styles.row}
          onPress={() => navigate("chat", { userId: c.userId, title: c.title })}
        >
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{c.title[0]}</Text>
          </View>
          <View style={styles.rowMeta}>
            <Text style={styles.rowTitle}>{c.title}</Text>
            <Text style={styles.rowPreview}>{c.preview}</Text>
          </View>
          <View style={[styles.dot, { backgroundColor: c.online ? "#16a34a" : "#94a3b8" }]} />
        </TouchableOpacity>
      ))}

      <View style={styles.navRow}>
        <TouchableOpacity style={styles.navBtn} onPress={() => navigate("voice")}>
          <Text style={styles.navText}>语音</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.navBtn} onPress={() => navigate("settings")}>
          <Text style={styles.navText}>设置</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, padding: 20, paddingTop: 60, backgroundColor: "#f8fafc" },
  title: { fontSize: 28, fontWeight: "700", color: "#0f172a", marginBottom: 8 },
  statusBadge: {
    alignSelf: "flex-start" as const, paddingHorizontal: 12, paddingVertical: 4,
    borderRadius: 12, marginBottom: 4,
  },
  statusText: { fontSize: 13, fontWeight: "600", color: "#334155" },
  error: { color: "#b91c1c", fontSize: 13, marginBottom: 8 },
  sectionTitle: { fontSize: 16, fontWeight: "600", color: "#0f172a", marginTop: 16, marginBottom: 8 },
  row: { flexDirection: "row", alignItems: "center", backgroundColor: "#fff", padding: 14, borderRadius: 12, marginBottom: 8 },
  avatar: { width: 40, height: 40, borderRadius: 20, backgroundColor: "#2f6bff", justifyContent: "center", alignItems: "center", marginRight: 12 },
  avatarText: { color: "#fff", fontSize: 16, fontWeight: "700" },
  rowMeta: { flex: 1 },
  rowTitle: { fontSize: 15, fontWeight: "600", color: "#0f172a" },
  rowPreview: { fontSize: 13, color: "#64748b", marginTop: 2 },
  dot: { width: 10, height: 10, borderRadius: 5 },
  navRow: { flexDirection: "row", gap: 12, marginTop: 20 },
  navBtn: { flex: 1, backgroundColor: "#2f6bff", paddingVertical: 12, borderRadius: 10, alignItems: "center" as const },
  navText: { color: "#fff", fontSize: 15, fontWeight: "600" },
});
