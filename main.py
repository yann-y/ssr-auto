import requests
import json

host = "127.0.0.1:9090"


def getNode(geturl):
    res = requests.get(geturl).json()
    json_str = json.dumps(res, sort_keys=True)
    # 将 JSON 对象转换为 Python 字典
    params_json = json.loads(json_str)
    items = params_json['proxies'].items()
    node_set = set()
    for key, value in items:
        str_key = str(key)
        if str_key[:2] == "V4":
            node_set.add(str_key)
    return node_set


def getTimeOut(geturl, node_set):
    result_map = {}
    timeout = 5000
    url = "http://www.gstatic.com/generate_204"
    for i in node_set:
        requests_url = geturl + "/" + i + "/delay?" + "timeout=" + str(timeout) + "&url=" + url
        res = requests.get(requests_url).json()
        dict = eval(str(res))
        ok = dict.get("delay", -1)
        if ok == -1:
            continue
        result_map[i] = dict["delay"]
        # print(res, '\n')
    return result_map


def postNode(geturl, result_mp):
    geturl = geturl + "/Proxy"
    f = zip(result_mp.values(), result_mp.keys())
    result_t = list(f)[0]
    # print(result_t)
    data = {"name": result_t[1]}
    # print(data)
    res = requests.put(url=geturl, json=data)
    print(res.status_code)
    if res.status_code == 204:
        print("--->>更换节点成功！！<<--- --》", result_t)
    else:
        print("--->>更换节点失败！！<<---")


if __name__ == '__main__':
    url = "http://127.0.0.1:9090/proxies"
    node_set = getNode(url)
    result_map = getTimeOut(url, node_set)
    postNode(url, result_map)
