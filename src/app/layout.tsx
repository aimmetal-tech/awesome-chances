import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "awesome-chances | AI 导学与成长规划",
  description: "面向校园学生和计算机自学初学者的 AI 导学、成长规划与真实任务推荐平台。",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="zh-CN" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
