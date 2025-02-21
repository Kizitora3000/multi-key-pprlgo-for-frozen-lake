import matplotlib.pyplot as plt
import pandas as pd

plt.rcParams["font.size"] = 28

# CSVファイルのパスを設定
mkpprl_5x5_1_agent_path = 'MKPPRL_average_success_rate_5x5_in_userNum_1.csv'
mkpprl_5x5_2_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_2.csv'
mkpprl_5x5_3_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_3.csv'
mkpprl_5x5_4_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_4.csv'
mkpprl_5x5_5_agents_path = 'MKPPRL_average_success_rate_5x5_in_userNum_5.csv'

# CSVファイルの読み込み
mkpprl_5x5_1_agent = pd.read_csv(mkpprl_5x5_1_agent_path)
mkpprl_5x5_2_agents = pd.read_csv(mkpprl_5x5_2_agents_path)
mkpprl_5x5_3_agents = pd.read_csv(mkpprl_5x5_3_agents_path)
mkpprl_5x5_4_agents = pd.read_csv(mkpprl_5x5_4_agents_path)
mkpprl_5x5_5_agents = pd.read_csv(mkpprl_5x5_5_agents_path)

# グラフの描画
plt.figure(figsize=(8, 6))

# 4x4 Configuration
h_pprl_5x5, = plt.plot(mkpprl_5x5_1_agent['Episode'], mkpprl_5x5_1_agent['Average Success Rate'], 
                       label='1 agent', color='blue', linestyle='-', linewidth=3)
h_pprl_5x5, = plt.plot(mkpprl_5x5_2_agents['Episode'], mkpprl_5x5_2_agents['Average Success Rate'], 
                       label='2 agents', color='red', linestyle='-', linewidth=3)
h_pprl_5x5, = plt.plot(mkpprl_5x5_3_agents['Episode'], mkpprl_5x5_3_agents['Average Success Rate'], 
                       label='3 agents', color='green', linestyle='-', linewidth=3)
h_pprl_5x5, = plt.plot(mkpprl_5x5_4_agents['Episode'], mkpprl_5x5_4_agents['Average Success Rate'], 
                       label='4 agents', color='yellow', linestyle='-', linewidth=3)
h_pprl_5x5, = plt.plot(mkpprl_5x5_5_agents['Episode'], mkpprl_5x5_5_agents['Average Success Rate'], 
                       label='5 agents', color='purple', linestyle='-', linewidth=3)


plt.xlabel('Episode')
plt.ylabel('Average Success Rate')
plt.legend(loc='lower right')
plt.grid(True)

plt.show()