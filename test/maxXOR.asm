# Test to find maximum XOR-value of at-most k-elements from 1 to n

	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
k_str:	.asciiz "Enter k: "
k:	.word	0
retVal:	.word	0
str:	.asciiz "Maximum XOR-value: "
x:	.word	0
result:	.word	0

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
	li $v0, 4
	la $a0, k_str
	syscall
	li $v0, 5
	syscall
	move $t4, $v0
	sw $t1, n
	sw $t4, k
	jal maxXOR
	lw $t1, n
	lw $t4, k
	sw $t1, n		# spilled n, freed $t1
	move $t1, $v0
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, retVal
	sw $t4, k
	li $v0, 10
	syscall
	.end main

	.globl maxXOR
	.ent maxXOR
maxXOR:
	# x = log2(n) + 1
	li $t1, 0		# x -> $t1

	# Store variables back into memory
	sw $t1, x

while:
	lw $t1, n
	srl $t1, $t1, 1		# n -> $t1
	lw $t4, x
	addi $t4, $t4, 1	# x -> $t4

	# Store variables back into memory
	sw $t1, n
	sw $t4, x

	lw $t1, n
	bgt $t1, 0, while		# while -> $t0
	# Return (2^x - 1)
	li $t4, 1		# result -> $t4
	sw $t1, n		# spilled n, freed $t1
	lw $t1, x
	sll $t4, $t4, $t1	# result -> $t4
	sub $t4, $t4, 1		# result -> $t4
	move $v0, $t4

	# Store variables back into memory
	sw $t1, x
	sw $t4, result

	jr $ra
	.end maxXOR
