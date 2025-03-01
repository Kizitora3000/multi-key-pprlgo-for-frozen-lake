import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.scale as mscale
import matplotlib.transforms as mtransforms
import matplotlib.ticker as ticker
import numpy as np

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
    
    plt.rcParams['font.family'] = 'MS Gothic'
    plt.rcParams['font.size'] = 32

    # カラーコードとマーカーの設定
    colors = {
        16: 'black',
        25: 'skyblue',
        36: '#5a4498'
    }
    
    markers = {
        16: 'o',
        25: '^',
        36: 's'
    }

    plt.figure(figsize=(8, 6))
    for state, time_list in times.items():
        plt.plot(num_people, time_list, 
                marker=markers[state], 
                color=colors[state],  
                label=f'$N_s$: {state}',
                markersize=12,       # マーカーサイズを設定
                linewidth=3)         # 線の太さを設定
    
    plt.xlabel("Number of users")
    plt.ylabel("Processing time [s]", labelpad=10)
    plt.legend(loc='lower right')
    plt.grid()

    # x軸とy軸の両方に整数値のみの目盛りを設定
    plt.gca().xaxis.set_major_locator(ticker.MaxNLocator(integer=True))
    plt.gca().yaxis.set_major_locator(ticker.MaxNLocator(integer=True))
    
    # 軸の範囲を設定
    plt.xlim(0.8, max(num_people) * 1.07)
    plt.ylim(0, max(max(time_list) for time_list in times.values()) * 1.1)
    
    plt.show()

file_path = "performance.csv"
plot_processing_time(file_path)