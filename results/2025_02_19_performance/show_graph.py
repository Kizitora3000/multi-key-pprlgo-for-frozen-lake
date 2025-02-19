import pandas as pd
import matplotlib.pyplot as plt

# CSVファイルの読み込み
def load_data(file_path):
    df = pd.read_csv(file_path)
    state_counts = df["状態数"].values
    num_people = df.columns[1:].astype(int)
    
    def time_to_seconds(time_str):
        if 'm' in time_str:
            minutes, seconds = time_str.split('m')
            return int(minutes) * 60 + float(seconds.replace('s', ''))
        else:
            return float(time_str.replace('s', ''))
    
    times = {state_counts[i]: [time_to_seconds(df.iloc[i, j]) for j in range(1, len(df.columns))] for i in range(len(state_counts))}
    return num_people, times

# グラフの作成
def plot_processing_time(file_path):
    num_people, times = load_data(file_path)
    
    # plt.rcParams['font.family'] = 'MS Gothic'
    plt.rcParams['font.size'] = 28

    plt.figure(figsize=(8, 6))
    for state, time_list in times.items():
        plt.plot(num_people, time_list, marker='o', label=f'$N_s$: {state}')
    
    plt.xlabel("Number of users")
    plt.ylabel("Processing time [seconds]")
    # plt.title("人数ごとの処理時間 (状態数別)")
    plt.legend(bbox_to_anchor=(1.05, 1), loc='upper left', borderaxespad=0.)
    plt.grid()
    plt.xticks(num_people)
    plt.show()

file_path = "performance.csv"
plot_processing_time(file_path)
