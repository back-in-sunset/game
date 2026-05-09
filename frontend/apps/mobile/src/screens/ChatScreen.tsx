import { useState } from "react";
import { View, Text, TextInput, FlatList, TouchableOpacity, StyleSheet } from "react-native";
import { useIMStore } from "../store/imStore";

type Props = {
  userId: number;
  title: string;
  onBack: () => void;
};

export function ChatScreen({ userId, title, onBack }: Props) {
  const { messages, sendMessage } = useIMStore();
  const [text, setText] = useState("");

  function handleSend() {
    if (!text.trim()) return;
    sendMessage(userId, text.trim());
    setText("");
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={onBack}>
          <Text style={styles.backBtn}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.title}>{title}</Text>
      </View>

      <FlatList
        data={messages}
        keyExtractor={(item) => item.id}
        style={styles.list}
        renderItem={({ item }) => (
          <View style={[styles.bubble, item.mine ? styles.mine : styles.other]}>
            <Text style={item.mine ? styles.mineText : styles.otherText}>{item.text}</Text>
            <Text style={styles.time}>{item.time}</Text>
          </View>
        )}
      />

      <View style={styles.composer}>
        <TextInput
          style={styles.input}
          placeholder="消息..."
          value={text}
          onChangeText={setText}
          onSubmitEditing={handleSend}
          returnKeyType="send"
        />
        <TouchableOpacity style={styles.sendBtn} onPress={handleSend}>
          <Text style={styles.sendText}>发送</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  header: { flexDirection: "row", alignItems: "center", padding: 16, paddingTop: 60, backgroundColor: "#fff", borderBottomWidth: 1, borderColor: "#e2e8f0" },
  backBtn: { fontSize: 15, color: "#2f6bff", marginRight: 12 },
  title: { fontSize: 18, fontWeight: "600", color: "#0f172a" },
  list: { flex: 1, padding: 16 },
  bubble: { maxWidth: "80%", padding: 12, borderRadius: 12, marginBottom: 8 },
  mine: { alignSelf: "flex-end", backgroundColor: "#2f6bff" },
  other: { alignSelf: "flex-start", backgroundColor: "#fff" },
  mineText: { color: "#fff", fontSize: 15 },
  otherText: { color: "#0f172a", fontSize: 15 },
  time: { fontSize: 11, color: "#94a3b8", marginTop: 4 },
  composer: { flexDirection: "row", alignItems: "center", padding: 12, backgroundColor: "#fff", borderTopWidth: 1, borderColor: "#e2e8f0" },
  input: { flex: 1, backgroundColor: "#f1f5f9", borderRadius: 10, paddingHorizontal: 14, paddingVertical: 10, fontSize: 15 },
  sendBtn: { marginLeft: 10, backgroundColor: "#2f6bff", paddingHorizontal: 16, paddingVertical: 10, borderRadius: 10 },
  sendText: { color: "#fff", fontWeight: "600", fontSize: 15 },
});
