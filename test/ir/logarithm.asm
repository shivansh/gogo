# Test to find floor value of logarithm (base 2 and base 10) of a number

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
base2Str:	.asciiz "log2(n): "
base10Str:	.asciiz "\nlog10(n): "

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
	li	$7, -1		# i -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, i

while:
	lw	$3, n
	srl	$3, $3, 1	# n -> $3
	lw	$7, i
	addi	$7, $7, 1	# i -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, i
	bgt	$3, 0, while	# while -> $0

	li	$v0, 4
	la	$a0, base2Str
	syscall
	li	$v0, 1
	lw	$3, i
	move	$a0, $3
	syscall
	li	$v0, 4
	la	$a0, base10Str
	syscall
	# Ideally
	div	$3, $3, 3	# i -> $3
	li	$v0, 1
	move	$a0, $3
	syscall
	# Store variables back into memory
	sw	$3, i
	li	$v0, 10
	syscall
	.end main
