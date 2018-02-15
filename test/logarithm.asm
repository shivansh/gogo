# Test to find floor value of logarithm (base 2 and base 10) of a number

	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
i:	.word	0
base2Str:	.asciiz "log2(n): "
base10Str:	.asciiz "\nlog10(n): "

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, -1		# i -> $t4

	# Store variables back into memory
	sw $t1, n
	sw $t4, i

while:
	lw $t1, n
	srl $t1, $t1, 1		# n -> $t1
	lw $t4, i
	addi $t4, $t4, 1	# i -> $t4

	# Store variables back into memory
	sw $t4, i
	sw $t1, n

	lw $t1, n
	bgt $t1, 0, while		# while -> $t0
	li $v0, 4
	la $a0, base2Str
	syscall
	li $v0, 1
	lw $t4, i
	move $a0, $t4
	syscall
	li $v0, 4
	la $a0, base10Str
	syscall
	# Ideally
	div $t4, $t4, 3		# i -> $t4
	li $v0, 1
	move $a0, $t4
	syscall

	# Store variables back into memory
	sw $t1, n
	sw $t4, i
	li $v0, 10
	syscall
	.end main
