import matplotlib.pyplot as plt
import pandas as pd

plt.rcParams["font.size"] = 44

# CSVファイルのパスを設定
mkpprl_5x5_1_agent_path = 'MKPPRL_average_success_rate_5x5_in_userNum_1.csv'
mkpprl_5x5_2_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_2.csv'
mkpprl_5x5_3_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_3.csv'

# CSVファイルの読み込み
mkpprl_5x5_1_agent = pd.read_csv(mkpprl_5x5_1_agent_path)
mkpprl_5x5_2_agents = pd.read_csv(mkpprl_5x5_2_agents_path)
mkpprl_5x5_3_agents = pd.read_csv(mkpprl_5x5_3_agents_path)
# グラフの描画
plt.figure(figsize=(8, 6))

linewidth = 5

# 4x4 Configuration
h_pprl_5x5, = plt.plot(mkpprl_5x5_1_agent['Episode'], mkpprl_5x5_1_agent['Average Success Rate'], 
                       label='1 agent', color='#708090', linestyle='-', linewidth=linewidth)
h_pprl_5x5, = plt.plot(mkpprl_5x5_2_agents['Episode'], mkpprl_5x5_2_agents['Average Success Rate'], 
                       label='2 agents', color='#FFD700', linestyle='-', linewidth=linewidth)
h_pprl_5x5, = plt.plot(mkpprl_5x5_3_agents['Episode'], mkpprl_5x5_3_agents['Average Success Rate'], 
                       label='3 agents', color='#40E0D0', linestyle='-', linewidth=linewidth)


plt.xlabel('Episode')
plt.ylabel('Average Success Rate')
plt.legend(loc='lower right')
plt.grid(True)

plt.subplots_adjust(top=0.995,bottom=0.13,left=0.09,right=0.995,hspace=0.2,
wspace=0.2)

plt.show()