import re, os

class SnapShot:
    DIMEX_State: int
    lcl: int
    reqTs: int
    nbrResps: int
    receivedResps: str
    waiting: list[bool]
    messages: list[str]

    def __init__(self, _d, _l, _rt, _n, _rr, _w:str, _m):
        self.DIMEX_State = _d
        self.lcl = _l
        self.reqTs = _rt
        self.nbrResps = _n
        self.receivedResps = _rr

        self.waiting = list(map(lambda x: x == 'true',_w.split(' ')))
        self.messages = _m.split(';')

regex = re.compile(
    r"DIMEX_State: ([0-2])(?:\s*)lcl: ([0-9]+)(?:\s*)reqTs: ([0-9]+)(?:\s*)nbrResps: ([0-2])(?:\s*)receivedResps: \[(.*)\](?:\s*)waiting: \[(\w+ \w+ \w+)\](?:\s*)messages in channels: \[(.*)\](?:\s*)",
)

def process(file) -> list[SnapShot]:
    snapshots = []
    with open(file) as fp:
        text = fp.read()
        l = re.findall(regex, text)
        #print(l)
        for snap in l:
            snapshots.append(SnapShot(int(snap[0]),int(snap[1]),int(snap[2]),int(snap[3]),snap[4],snap[5],snap[6]))
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

# inv 2: se todos processos estão em "não quero a SC", então todos waitings tem que ser falsos e não deve haver mensagens
def test_inv2(s0,s1,s2):
    for i in range(len(s0)):
        count = 0
        if s0[i].DIMEX_State == 0:
            count += 1
        if s1[i].DIMEX_State == 0:
            count += 1
        if s2[i].DIMEX_State == 0:
            count += 1
        if count == 3:
            print("Aconteceu o caso legal da inv 2")
            if s0[i].waiting != [False, False, False] or s1[i].waiting != [False, False, False] or s2[i].waiting != [False, False, False]:
                print("Violação da inv 2 por waiting")
            # Acho que mensagem do snapshot pode ativar esse caso...
            if len(s0[i].messages) > 0 or len(s1[i].messages) > 0 or len(s2[i].messages) > 0:
                print("Violação da inv 2 por mensagens")

#inv 3: se um processo q está marcado como waiting em p, então p está na SC ou quer a SC
def test_inv3(s:list[list[SnapShot]]):
    for i in range(len(s0)):
        for p in range(3):
            if s[0][i].waiting[p]:
                if s[0][i].DIMEX_State == 0:
                    print(f"Violação da inv 3 na snap {i} com o waiting do 0 discordando do DIMEX_State do {p}")

            if s[1][i].waiting[p]:
                if s[1][i].DIMEX_State == 0:
                    print(f"Violação da inv 3 na snap {i} com o waiting do 1 discordando do DIMEX_State do {p}")

            if s[2][i].waiting[p]:
                if s[2][i].DIMEX_State == 0:
                    print(f"Violação da inv 3 na snap {i} com o waiting do 2 discordando do DIMEX_State do {p}")

#inv 4: se um processo q quer a seção crítica (nao entrou ainda),
#       então o somatório de mensagens recebidas, de mensagens em transito e de, flags waiting para p 
#       em outros processos deve ser igual a N-1  (onde N é o número total de processos)
def test_inv4(s0,s1,s2):
    for i in range(len(s0)):
        if s0[i].DIMEX_State == 1:
            n = inv4_internal(s0[i],s1[i],s2[i],0)
            if n != 2:
                print(f"Violação da inv 4 na snap {i}, {n} em vez de 2")
        if s1[i].DIMEX_State == 1:
            n = inv4_internal(s1[i],s0[i],s2[i],1)
            if n != 2:
                print(f"Violação da inv 4 na snap {i}, {n} em vez de 2")
        if s2[i].DIMEX_State == 1:
            n = inv4_internal(s2[i],s1[i],s0[i],2)
            if n != 2:
                print(f"Violação da inv 4 na snap {i}, {n} em vez de 2")
                
# o somatório de mensagens recebidas, de mensagens em transito 
# e de flags waiting para p em outros processos
def inv4_internal(target,other1,other2, target_id):
   sum = target.nbrResps
   if other1.waiting[target_id]:
       sum += 1
   if other2.waiting[target_id]:
       sum += 1
   for i in range(3):
       if i == target_id:
           continue
       if f"respOk {i};" in target.messages:
           sum += 1
   return sum
   # return target.nbrResps + 1 if other1.waiting[target_id] else 0 + 1 if other2.waiting[target_id] else 0

s0 = process("snapshots_0.txt")
s1 = process("snapshots_1.txt")
s2 = process("snapshots_2.txt")


test_dois_processos_na_secao_critica(s0,s1,s2)
test_inv2(s0,s1,s2)
test_inv3([s0,s1,s2])
test_inv4(s0,s1,s2)