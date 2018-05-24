# Return (2^x - 1)

# Test to find maximum XOR-value of at-most k-elements from 1 to n

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
kStr:		.asciiz "Enter k: "
k:		.word	0
retVal:		.word	0
str:		.asciiz "Maximum XOR-value: "
x:		.word	0
result:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, nStr
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	li	$v0, 4
	la	$a0, kStr
	syscall
	li	$v0, 5
	syscall
	move	$7, $v0
	sw	$3, n
	sw	$7, k
	jal	maxXOR
	lw	$3, n
	lw	$7, k
	move	$15, $v0
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 1
	move	$a0, $15
	syscall
	# Store variables back into memory
	sw	$3, n
	sw	$7, k
	sw	$15, retVal
	li	$v0, 10
	syscall
	.end main
	.globl maxXOR
	.ent maxXOR
maxXOR:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	# x = log2(n) + 1
	li	$3, 0		# x -> $3
	# Store variables back into memory
	sw	$3, x

while:
	lw	$3, n
	srl	$3, $3, 1	# n -> $3
	lw	$7, x
	addi	$7, $7, 1	# x -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, x
	bgt	$3, 0, while	# while -> $0

	li	$3, 1		# result -> $3
	lw	$7, x
	sll	$3, $3, $7	# result -> $3
	sub	$3, $3, 1	# result -> $3
	move	$v0, $3
	# Store variables back into memory
	sw	$3, result
	sw	$7, x

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end maxXOR
