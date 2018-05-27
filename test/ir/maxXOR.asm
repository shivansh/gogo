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
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	li	$2, 4
	la	$4, kStr
	syscall
	li	$2, 5
	syscall
	sw	$3, n		# spilled n, freed $3
	move	$3, $2
	sw	$3, k
	jal	maxXOR
	lw	$3, k
	move	$3, $2
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, retVal
	li	$2, 10
	syscall
	.end main

	.globl maxXOR
	.ent maxXOR
maxXOR:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	# x = log2(n) + 1
	li	$3, 0		# x -> $3
	# Store dirty variables back into memory
	sw	$3, x

while:
	lw	$3, n
	srl	$3, $3, 1	# n -> $3
	lw	$5, x
	addi	$5, $5, 1	# x -> $5
	# Store dirty variables back into memory
	sw	$3, n
	sw	$5, x
	bgt	$3, 0, while

	li	$3, 1		# result -> $3
	lw	$5, x
	sll	$3, $3, $5	# result -> $3
	sub	$3, $3, 1	# result -> $3
	move	$2, $3
	# Store dirty variables back into memory
	sw	$3, result

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end maxXOR
