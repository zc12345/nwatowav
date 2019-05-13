# coding=utf8
import os
import subprocess

def batch(input_dir, output_dir):
    ext_dict = {"nwa":"wav", "nwk":"wav", "ovk":"ogg"}
    CREATE_NO_WINDOW = 0x08000000
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    for root, dirs, files in os.walk(input_dir):
        n = len(files)
        l = min(n, 20)
        for i, f in enumerate(files):
            if f[-3:] in ext_dict:
                inputfile = os.path.join(input_dir, f)
                outputfile = os.path.join(output_dir, f[:-4])
                cmd = "./nwatowav --inputfile=\"" + inputfile + "\" --outputfile=\"" + outputfile + "\""
                subprocess.call(cmd, creationflags=CREATE_NO_WINDOW)
            finish = round((i + 1)*l / n)
            remain = l - finish
            msg = "#"*finish + "="*remain + "process:{0}%".format(round((i + 1)*100 / n))
            print(msg)


if __name__ == "__main__":
    input_dir = "../nwatowav_o/wav"
    output_dir = "./wav"
    if not os.path.isfile("nwatowav.exe"):
        subprocess.Popen("go get github.com/hasenbanck/nwa", shell=True)
        subprocess.Popen("go build", shell=True)
        print("build over")
    batch(input_dir, output_dir)
