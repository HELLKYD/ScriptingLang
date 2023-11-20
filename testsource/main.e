fun main() int {
    let x int = 4*1;
    let y int = x;
    x += test2(1 + 1) + y;
    println(x);
    return x;
}

fun test(x int, y int, z int) int {
    x += y;
    return x + z;
}

fun test2(x int) int {
    return x;
}