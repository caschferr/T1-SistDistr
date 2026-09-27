import re, os

class SnapShot:
    DIMEX_State: int
    lcl: int
    reqTs: int
    nbrResps: int
    receivedResps: str
    waiting: str
    messages: str # ele não lê isso por enquanto

    def __init__(self, _d, _l, _rt, _n, _rr, _w):
        self.DIMEX_State = _d
        self.lcl = _l
        self.reqTs = _rt
        self.nbrResps = _n
        self.receivedResps = _rr
        self.waiting = _w

regex = re.compile(
    r"DIMEX_State: ([0-2])(?:\s*)lcl: ([0-9]+)(?:\s*)reqTs: ([0-9]+)(?:\s*)nbrResps: ([0-2])(?:\s*)receivedResps: \[(.*)\](?:\s*)waiting: (\w+ \w+ \w+)(?:\s*)messages in channels:(?:\s*)",
)

def process(file) -> list[SnapShot]:
    snapshots = []
    with open(file) as fp:
        text = fp.read()
        l = re.findall(regex, text)
        #print(l)
        for snap in l:
            snapshots.append(SnapShot(int(snap[0]),int(snap[1]),int(snap[2]),int(snap[3]),snap[4],snap[5]))
        return snapshots

def test_dois_processos_na_secao_critica(s0,s1,s2):
    for i in range(len(s0)):
        count = 0
        if s0[i].DIMEX_State == 2:
            count += 1
        if s1[i].DIMEX_State == 2:
            count += 1
        if s2[i].DIMEX_State == 2:
            count += 1
        if count > 1:
            print("Mais de um processo na seção crítica")

s0 = process("snapshots_0.txt")
s1 = process("snapshots_1.txt")
s2 = process("snapshots_2.txt")

test_dois_processos_na_secao_critica(s0,s1,s2)