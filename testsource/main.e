fun main() int {
    let x int = 4*1;
    let y int = x;
    x += y;
    println(x);
    return x;
}

fun test(x int) {
    x += 2;
    x = 2*x;
    return x;
}