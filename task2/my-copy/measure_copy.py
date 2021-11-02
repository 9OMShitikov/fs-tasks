import time
import subprocess

mean_cnt = 7

N = [4, 8]
read_size = [(256 * 1024, "256K"),
             (512 * 1024, "512K")]


def test_on_sizes(file_size, file_name):
    import os
    with open(file_name, 'wb') as f_out:
        f_out.write(os.urandom(file_size))

    f_out.close()

    result = file_name + "," + str(file_size) + ","
    for n in N:
        for r_size, _ in read_size:
            start = time.time()
            for _ in range(mean_cnt):
                subprocess.run(["./my_copy", file_name, file_name + "_copy", str(n), str(r_size)])
            finish = time.time()
            result += str((finish - start) / mean_cnt) + ","

    start = time.time()
    for _ in range(mean_cnt):
        subprocess.run(["cp", file_name, file_name + "_copy"])
    finish = time.time()
    result += str((finish - start) / mean_cnt) + "\n"
    os.remove(file_name)
    os.remove(file_name + "_copy")
    return result


results = open("results.csv", 'w')
f_sizes = open("file_sizes.txt", "r")
names = "size_name,size_val,"
for n in N:
    for _, r_name in read_size:
        names += str(n) + " запроса по " + r_name + ","

names += "cp\n"
results.write(names)
for line in f_sizes:
    size, name = tuple(line.split(","))
    size = int(size)
    result = test_on_sizes(size, name[1:-1])
    results.write(result)

results.close()
f_sizes.close()

