import matplotlib.pyplot as plt
import pandas as pd

# CSVファイルからデータを読み込む
df1 = pd.read_csv('MKPPRL_success_rate_4x4_in_usernum_1.csv')
df2 = pd.read_csv('MKPPRL_success_rate_4x4_in_usernum_2.csv')
df3 = pd.read_csv('MKPPRL_success_rate_4x4_in_usernum_3.csv')

# グラフを描画
plt.figure(figsize=(10, 6))

plt.plot(df1['Episode'], df1['Success Rate'], label='1 agent', linestyle='-', color='blue')
plt.plot(df2['Episode'], df2['Success Rate'], label='2 agents', linestyle='--', color='red')
plt.plot(df3['Episode'], df3['Success Rate'], label='3 agents', linestyle='-.', color='green')

plt.title('4x4')
plt.xlabel('Episode')
plt.ylabel('Success Rate')
plt.legend()
plt.grid(True)

plt.show()