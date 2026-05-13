import { BrowserRouter, Route, Routes } from "react-router-dom";
import { PlatformConsolePage } from "../pages/PlatformConsolePage";

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="*" element={<PlatformConsolePage />} />
      </Routes>
    </BrowserRouter>
  );
}
