import matplotlib.pyplot as plt
import pandas as pd

# CSVファイルのパスを設定
mkpprl_4x4_path = 'MKPPRL_average_success_rate_4x4.csv'
pprl_4x4_path = 'PPRL_average_success_rate_4x4.csv'
mkpprl_5x5_path = 'MKPPRL_average_success_rate_5x5.csv'
pprl_5x5_path = 'PPRL_average_success_rate_5x5.csv'
mkpprl_6x6_path = 'MKPPRL_average_success_rate_6x6.csv'
pprl_6x6_path = 'PPRL_average_success_rate_6x6.csv'

# CSVファイルの読み込み
mkpprl_4x4_df = pd.read_csv(mkpprl_4x4_path)
pprl_4x4_df = pd.read_csv(pprl_4x4_path)
mkpprl_5x5_df = pd.read_csv(mkpprl_5x5_path)
pprl_5x5_df = pd.read_csv(pprl_5x5_path)
mkpprl_6x6_df = pd.read_csv(mkpprl_6x6_path)
pprl_6x6_df = pd.read_csv(pprl_6x6_path)

# グラフの描画
plt.figure(figsize=(10, 6))

# 4x4 Configuration
plt.plot(mkpprl_4x4_df['Episode'], mkpprl_4x4_df['Average Success Rate'], 
         label='MKPPRL 4x4', color='darkblue', linestyle='-', linewidth=2)
plt.plot(pprl_4x4_df['Episode'], pprl_4x4_df['Average Success Rate'], 
         label='PPRL 4x4', color='lightblue', linestyle=':', linewidth=2)

# 5x5 Configuration
plt.plot(mkpprl_5x5_df['Episode'], mkpprl_5x5_df['Average Success Rate'], 
         label='MKPPRL 5x5', color='darkred', linestyle='-', linewidth=2)
plt.plot(pprl_5x5_df['Episode'], pprl_5x5_df['Average Success Rate'], 
         label='PPRL 5x5', color='lightcoral', linestyle=':', linewidth=2)

# 6x6 Configuration
plt.plot(mkpprl_6x6_df['Episode'], mkpprl_6x6_df['Average Success Rate'], 
         label='MKPPRL 6x6', color='darkgreen', linestyle='-', linewidth=2)
plt.plot(pprl_6x6_df['Episode'], pprl_6x6_df['Average Success Rate'], 
         label='PPRL 6x6', color='lightgreen', linestyle=':', linewidth=2)

plt.xlabel('Episode')
plt.ylabel('Average Success Rate')
plt.legend()
plt.grid(True)

plt.show()